package util

import (
	"log"
	"os"
	"os/exec"
)

// SetupNginxUser creates the nginx group and user and assigns ownership to nginx directories
func SetupNginxUser() {
	err := exec.Command("useradd", "--shell", "/bin/false", "--home", "/dev/null", "nginx").Run()

	if err != nil {
		log.Printf("Failed creating 'nginx' user. Does he already exist? Error: %s", err)
	}
}

// SetupSystemd sets up the recommended nginx.service file
func SetupSystemd() {
	CopyFile("files/nginx.service", "/lib/systemd/system/nginx.service")

	// reload systemd daemon
	err := exec.Command("systemctl", "daemon-reload").Run()
	if err != nil {
		log.Printf("Failed reloading systemctl daemon Error: %s", err)
	}
}

// SetupInitD downloads the recommended LSB compliant init.d script for NginX
func SetupInitD() {
	DownloadFile("https://raw.githubusercontent.com/Fleshgrinder/nginx-sysvinit-script/master/init", "/etc/init.d/nginx")
	err := exec.Command("chmod", "+x", "/etc/init.d/nginx").Run()
	if err != nil {
		log.Printf("Failed assigning write privilege to /etc/init.d/nginx Error: %s", err)
	}
}

// SetupFileStructure creates all required folder for NginX and moves the delivered nginx file structure to /etc/nginx
func SetupFileStructure() {
	err := os.MkdirAll("/var/www/", os.ModeDir)

	// only check for error once, because it's most probable a problem of missing access rights
	if err != nil {
		log.Printf("Failed creating /var/www/ directory Error: %s", err)
	}

	os.MkdirAll("/var/cache/nginx", os.ModeDir)
	os.MkdirAll("/var/log/nginx", os.ModeDir)
	err = exec.Command("mv", "/etc/nginx", "/etc/nginx-default").Run()

	// copy our nginx template folder to /etc/nginx
	err = exec.Command("cp", "-r", "nginx", "/etc/nginx").Run()

	if err != nil {
		log.Printf("Failed copying 'nginx' folder to /etc/nginx Error: %s", err)
	}
}

// GenerateDHParams for NginX DHE key exchange (strength of 4096bit)
// using -dsaparam to disable prime number check => speedup (but not less secure!)
func GenerateDHParams() {
	err := exec.Command("openssl", "dhparam", "-dsaparam", "-out", "/etc/nginx/ssl/dhparam.pem", "4096").Run()

	if err != nil {
		log.Printf("Failed creating OpenSSL DHParam Error: %s", err)
	}
}
