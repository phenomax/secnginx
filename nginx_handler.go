package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/mholt/archiver"
	"github.com/phenomax/secnginx/util"
	"github.com/urfave/cli"
)

// CLIOptions wrapper for the parsed cli flags
type CLIOptions struct {
	Brotli, CORS, DynamicTLS, HeadersMore, CookieFlag, CT, Upgrade bool
}

const brotliModulePath string = "/build/modules/ngx_brotli"
const corsModulePath string = "/build/modules/ngx_http_cors_filter"
const headersMoreModulePath string = "/build/modules/headers-more-nginx-module"
const ctModulePath string = "/build/modules/nginx-ct"
const cookieFlagModulePath string = "/build/modules/nginx_cookie_flag_module"

const nginxPath = "/build/nginx"
const pcrePath = "/build/pcre"
const zlibPath = "/build/zlib"
const openSSLPath = "/build/openssl"

// Count - returns the amount of positive cli options in order to determine the amount of download goroutines
func (cli CLIOptions) Count() int {
	count := 0
	if cli.Brotli {
		count++
	}
	if cli.CORS {
		count++
	}
	if cli.HeadersMore {
		count++
	}
	if cli.CT {
		count++
	}
	if cli.CookieFlag {
		count++
	}

	return count
}

func start(c *cli.Context) error {
	config, err := util.GetConfig()

	if err != nil {
		log.Fatalf("Fatal error reading config file: %s \n", err)
		return nil
	}

	wd, err := os.Getwd()
	// set working directory as environment variable
	os.Setenv("WD", wd)

	if err != nil {
		log.Fatalf("Can't determine working directory: %s \n", err)
	}

	cliOptions := &CLIOptions{
		!c.Bool("without-brotli-module"),
		!c.Bool("without-cors-module"),
		!c.Bool("without-dynamic-tls-module"),
		!c.Bool("without-headers-more-module"),
		!c.Bool("without-cookie-flag-module"),
		!c.Bool("without-ct-module"),
		c.Bool("upgrade"),
	}

	installRequiredPackages()
	loadDependencies(config, cliOptions, wd)
	configureNginX(config, cliOptions, wd)

	if cliOptions.DynamicTLS {
		applyDynamicTLSRecordsPatch(wd)
	}

	makeAndInstallNginX(wd)

	if !cliOptions.Upgrade {
		log.Println("Setting up NginX user")
		util.SetupNginxUser()
		log.Println("Setting up Init.D script")
		util.SetupInitD()
		log.Println("Setting up SystemD")
		util.SetupSystemd()
		log.Println("Setting up NginX file structure")
		util.SetupFileStructure()
		log.Println("Generating strong DHParams")
		util.GenerateDHParams()

		log.Println("\nNginX successfully installed! Run 'service nginx start' to start it.")
		log.Println("Don't forget to check the further steps, described in the README.me in order to deploy a secure NginX installation!")
	} else {
		log.Println("\nNginX successfully upgraded! Run 'service nginx restart' to start the new version.")
		log.Println("Don't forget to check the further steps, described in the README.me in order to deploy a secure NginX installation!")
	}

	return nil
}

func installRequiredPackages() {
	log.Println("Installing dependencies via apt")

	exec.Command("apt", "update").Run()
	exec.Command("apt", "install", "-y", "build-essential", "cmake", "git",
	"libpcre3-dev", "curl", "libcurl4-openssl-dev", "zlib1g-dev", "automake").Run()
}

func loadDependencies(config *util.Config, cliOptions *CLIOptions, wd string) {
	// clear previous builds
	os.RemoveAll(wd + "/build/")
	os.MkdirAll(wd+"/build/modules", os.ModePerm)

	var wg sync.WaitGroup
	var err error
	wg.Add(4 + cliOptions.Count())

	// download nginx
	go func() {
		log.Println("Downloading NginX version " + config.NginXVersion)
		err = util.DownloadFile(fmt.Sprintf("https://nginx.org/download/nginx-%s.tar.gz", config.NginXVersion), "build/nginx.tar.gz")

		if err != nil {
			log.Fatalf("Failed downloading selected NginX version! Is the version number valid?\n Error: %s", err)
			os.Exit(0)
		}

		// extract archive
		log.Println("Extracting NginX")
		err = archiver.TarGz.Open("build/nginx.tar.gz", "build")
		os.Rename("build/nginx-"+config.NginXVersion, "build/nginx")

		if err != nil {
			log.Fatalf("Failed extracting NginX!\n Error: %s", err)
			os.Exit(0)
		}

		wg.Done()
	}()

	// download pcre
	go func() {
		log.Println("Downloading PCRE version " + config.PCREVersion)
		err = util.DownloadFile(fmt.Sprintf("https://ftp.pcre.org/pub/pcre/pcre-%s.tar.gz", config.PCREVersion), "build/pcre.tar.gz")

		if err != nil {
			log.Fatalf("Failed downloading selected PCRE version! Is the version number valid?\n Error: %s", err)
			os.Exit(0)
		}

		// extract archive
		log.Println("Extracting PCRE")
		err = archiver.TarGz.Open("build/pcre.tar.gz", "build")
		os.Rename("build/pcre-"+config.PCREVersion, "build/pcre")

		if err != nil {
			log.Fatalf("Failed extracting PCRE!\n Error: %s", err)
			os.Exit(0)
		}

		// fix for "aclocal-1.15: command not found" error on NginX make
		cmd := exec.Command("autoreconf", "-f", "-i")
		cmd.Dir = wd + pcrePath
		cmd.Run()

		wg.Done()
	}()

	// download zlib
	go func() {
		log.Println("Downloading ZLib version " + config.ZLibVersion)
		err = util.DownloadFile(fmt.Sprintf("https://zlib.net/zlib-%s.tar.gz", config.ZLibVersion), "build/zlib.tar.gz")

		if err != nil {
			log.Fatalf("Failed downloading selected ZLib version! Is the version number valid?\n Error: %s", err)
			os.Exit(0)
		}

		// extract archive
		log.Println("Extracting ZLib")
		err = archiver.TarGz.Open("build/zlib.tar.gz", "build")
		os.Rename("build/zlib-"+config.ZLibVersion, "build/zlib")

		if err != nil {
			log.Fatalf("Failed extracting ZLib!\n Error: %s", err)
			os.Exit(0)
		}

		wg.Done()
	}()

	// download openssl
	go func() {
		log.Println("Downloading OpenSSL version " + config.OpenSSLVersion)
		err = util.DownloadFile(fmt.Sprintf("https://www.openssl.org/source/openssl-%s.tar.gz", config.OpenSSLVersion), "build/openssl.tar.gz")

		if err != nil {
			log.Fatalf("Failed downloading OpenSSL ZLib version! Is the version number valid?\n Error: %s", err)
			os.Exit(0)
		}

		// extract archive
		log.Println("Extracting OpenSSL")
		err := archiver.TarGz.Open("build/openssl.tar.gz", "build")
		os.Rename("build/openssl-"+config.OpenSSLVersion, "build/openssl")

		if err != nil {
			log.Fatalf("Failed extracting OpenSSL!\n Error: %s", err)
			os.Exit(0)
		}

		wg.Done()
	}()

	if cliOptions.Brotli {
		go func() {
			log.Println("Downloading ngx_brotli module")
			cmd := exec.Command("git", "clone", "https://github.com/google/ngx_brotli.git")
			cmd.Dir = wd + "/build/modules/"
			err := cmd.Run()

			if err != nil {
				log.Fatalf("Failed downloading ngx_brotli!\n Error: %s", err)
				os.Exit(0)
			}

			cmd = exec.Command("git", "submodule", "update", "--init")
			cmd.Dir = wd + brotliModulePath
			err = cmd.Run()

			if err != nil {
				log.Fatalf("Failed updating submodules for ngx_brotli!\n Error: %s", err)
				os.Exit(0)
			}

			wg.Done()
		}()
	}

	if cliOptions.CORS {
		go func() {
			log.Println("Downloading ngx_http_cors_filter module")
			cmd := exec.Command("git", "clone", "https://github.com/x-v8/ngx_http_cors_filter.git")
			cmd.Dir = wd + "/build/modules/"
			err := cmd.Run()

			if err != nil {
				log.Fatalf("Failed downloading ngx_http_cors_filter!\n Error: %s", err)
				os.Exit(0)
			}

			wg.Done()
		}()
	}

	if cliOptions.HeadersMore {
		go func() {
			log.Println("Downloading headers_more-nginx module")
			cmd := exec.Command("git", "clone", "https://github.com/openresty/headers-more-nginx-module.git")
			cmd.Dir = wd + "/build/modules/"
			err := cmd.Run()

			if err != nil {
				log.Fatalf("Failed downloading headers_more_nginx module!\n Error: %s", err)
				os.Exit(0)
			}

			wg.Done()
		}()
	}

	if cliOptions.CT {
		go func() {
			log.Println("Downloading nginx_ct module")
			cmd := exec.Command("git", "clone", "https://github.com/grahamedgecombe/nginx-ct.git")
			cmd.Dir = wd + "/build/modules/"
			err := cmd.Run()

			if err != nil {
				log.Fatalf("Failed downloading nginx_ct!\n Error: %s", err)
				os.Exit(0)
			}

			wg.Done()
		}()
	}

	if cliOptions.CookieFlag {
		go func() {
			log.Println("Downloading cookie_flag module")
			cmd := exec.Command("git", "clone", "https://github.com/AirisX/nginx_cookie_flag_module.git")
			cmd.Dir = wd + "/build/modules/"
			err := cmd.Run()

			if err != nil {
				log.Fatalf("Failed downloading cookie_flag!\n Error: %s", err)
				os.Exit(0)
			}

			wg.Done()
		}()
	}

	wg.Wait()
	if err != nil {
		log.Fatalf("Fatal error while downloading files: %s", err)
	}
}

func configureNginX(config *util.Config, cliOptions *CLIOptions, wd string) {
	log.Println("Configuring NginX\n")

	// configure NginX with the specified parameters
	configParams := append(regexp.MustCompile("\n").Split(config.Configuration, -1), regexp.MustCompile("\n").Split(config.Modules, -1)...)

	// search for config flags, we will take care of
	for i := 0; i < len(configParams); i++ {
		e := configParams[i]
		// delete empty elements
		if e == "" {
			configParams = append(configParams[:i], configParams[i+1:]...)
			// form the remove item index to start iterate next item
			i--
			continue
		}

		if strings.Contains(e, "--with-openssl=") || strings.Contains(e, "--with-pcre=") || strings.Contains(e, " --with-zlib=") {
			log.Fatalf("Illegal config parameter %s found! Please remove, because SecNginX will set it.", e)
		}
		configParams[i] = strings.Replace(e, " ", "", -1)
	}

	if cliOptions.Brotli {
		configParams = append(configParams, "--add-module="+wd+brotliModulePath)
	}
	if cliOptions.CORS {
		configParams = append(configParams, "--add-module="+wd+corsModulePath)
	}
	if cliOptions.HeadersMore {
		configParams = append(configParams, "--add-module="+wd+headersMoreModulePath)
	}
	if cliOptions.CT {
		configParams = append(configParams, "--add-module="+wd+ctModulePath)
	}
	if cliOptions.CookieFlag {
		configParams = append(configParams, "--add-module="+wd+cookieFlagModulePath)
	}

	// append OpenSSL, PCRE and ZLib
	configParams = append(configParams, "--with-openssl="+wd+openSSLPath, "--with-pcre="+wd+pcrePath, "--with-zlib="+wd+zlibPath)

	cmd := exec.Command("./configure", configParams...)
	cmd.Dir = wd + nginxPath
	util.RunAndPrintCommandOutput(cmd)
}

func applyDynamicTLSRecordsPatch(wd string) {
	log.Println("Applying Dynamic TLS Records patch to NginX")

	cmd := exec.Command("patch", "-p1", fmt.Sprintf("<%s/files/NginX-Dynamic-TLS-Records.patch", wd))
	cmd.Dir = wd + nginxPath
	err := cmd.Run()

	if err != nil {
		log.Fatalf("Fatal error while applying Dynamic TLS Records patch to NginX: %s", err)
	}
}

func makeAndInstallNginX(wd string) {
	log.Println("Running 'make' NginX")

	cmd := exec.Command("make")
	cmd.Dir = wd + nginxPath
	util.RunAndPrintCommandOutput(cmd)

	log.Println("Running 'make install' NginX")

	cmd = exec.Command("make", "install")
	cmd.Dir = wd + nginxPath
	util.RunAndPrintCommandOutput(cmd)
}
