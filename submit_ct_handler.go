package main

import (
	_ "crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/phenomax/secnginx/util"
	"github.com/urfave/cli"
)

const chromeCTLogsURL = "https://www.gstatic.com/ct/log_list/log_list.json"

var ctLogList = []util.CTLogProvider{
	util.CTLogProvider{
		"google_argon2018",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE0gBVBa3VR7QZu82V+ynXWD14JM3ORp37MtRxTmACJV5ZPtfUA7htQ2hofuigZQs+bnFZkje+qejxoyvk2Q1VaA==",
		"https://ct.googleapis.com/logs/argon2018/",
	},
	util.CTLogProvider{
		"google_argon2019",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEI3MQm+HzXvaYa2mVlhB4zknbtAT8cSxakmBoJcBKGqGwYS0bhxSpuvABM1kdBTDpQhXnVdcq+LSiukXJRpGHVg==",
		"https://ct.googleapis.com/logs/argon2019/",
	},
	util.CTLogProvider{
		"google_xenon2018",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1syJvwQdrv0a8dM2VAnK/SmHJNw/+FxC+CncFcnXMX2jNH9Xs7Q56FiV3taG5G2CokMsizhpcm7xXzuR3IHmag==",
		"https://ct.googleapis.com/logs/xenon2018/",
	},
	util.CTLogProvider{
		"google_xenon2019",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/XyDwqzXL9i2GTjMYkqaEyiRL0Dy9sHq/BTebFdshbvCaXXEh6mjUK0Yy+AsDcI4MpzF1l7Kded2MD5zi420gA==",
		"https://ct.googleapis.com/logs/xenon2019/",
	},
	util.CTLogProvider{
		"google_icarus",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETtK8v7MICve56qTHHDhhBOuV4IlUaESxZryCfk9QbG9co/CqPvTsgPDbCpp6oFtyAHwlDhnvr7JijXRD9Cb2FA==",
		"https://ct.googleapis.com/icarus/",
	},
	util.CTLogProvider{
		"google_pilot",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETtK8v7MICve56qTHHDhhBOuV4IlUaESxZryCfk9QbG9co/CqPvTsgPDbCpp6oFtyAHwlDhnvr7JijXRD9Cb2FA==",
		"https://ct.googleapis.com/pilot/",
	},
	util.CTLogProvider{
		"google_rocketeer",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfahLEimAoz2t01p3uMziiLOl/fHTDM0YDOhBRuiBARsV4UvxG2LdNgoIGLrtCzWE0J5APC2em4JlvR8EEEFMoA==",
		"https://ct.googleapis.com/rocketeer/",
	},
	util.CTLogProvider{
		"cloudflare_nimbus_2018",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEAsVpWvrH3Ke0VRaMg9ZQoQjb5g/xh1z3DDa6IuxY5DyPsk6brlvrUNXZzoIg0DcvFiAn2kd6xmu4Obk5XA/nRg==",
		"https://ct.cloudflare.com/logs/nimbus2018/",
	},
	util.CTLogProvider{
		"cloudflare_nimbus_2019",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEkZHz1v5r8a9LmXSMegYZAg4UW+Ug56GtNfJTDNFZuubEJYgWf4FcC5D+ZkYwttXTDSo4OkanG9b3AI4swIQ28g==",
		"https://ct.cloudflare.com/logs/nimbus2019/",
	},
	util.CTLogProvider{
		"digicert_server_2",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEzF05L2a4TH/BLgOhNKPoioYCrkoRxvcmajeb8Dj4XQmNY+gxa4Zmz3mzJTwe33i0qMVp+rfwgnliQ/bM/oFmhA==",
		"https://ct2.digicert-ct.com/log/",
	},
	util.CTLogProvider{
		"digicert_yeti_2018",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESYlKFDLLFmA9JScaiaNnqlU8oWDytxIYMfswHy9Esg0aiX+WnP/yj4O0ViEHtLwbmOQeSWBGkIu9YK9CLeer+g==",
		"https://yeti2018.ct.digicert.com/log/",
	},
	util.CTLogProvider{
		"digicert_yeti_2019",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEkZd/ow8X+FSVWAVSf8xzkFohcPph/x6pS1JHh7g1wnCZ5y/8Hk6jzJxs6t3YMAWz2CPd4VkCdxwKexGhcFxD9A==",
		"https://yeti2019.ct.digicert.com/log/",
	},
	util.CTLogProvider{
		"digicert_nessie_2018",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEVqpLa2W+Rz1XDZPBIyKJO+KKFOYZTj9MpJWnZeFUqzc5aivOiWEVhs8Gy2AlH3irWPFjIZPZMs3Dv7M+0LbPyQ==",
		"https://nessie2018.ct.digicert.com/log/",
	},
	util.CTLogProvider{
		"digicert_nessie_2019",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEX+0nudCKImd7QCtelhMrDW0OXni5RE10tiiClZesmrwUk2iHLCoTHHVV+yg5D4n/rxCRVyRhikPpVDOLMLxJaA==",
		"https://nessie2019.ct.digicert.com/log/",
	},
	util.CTLogProvider{
		"commodo_mammoth",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7+R9dC4VFbbpuyOL+yy14ceAmEf7QGlo/EmtYU6DRzwat43f/3swtLr/L8ugFOOt1YU/RFmMjGCL17ixv66MZw==",
		"https://mammoth.ct.comodo.com/",
	},
	util.CTLogProvider{
		"commodo_sabre",
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8m/SiQ8/xfiHHqtls9m7FyOMBg4JVZY9CgiixXGz0akvKD6DEL8S0ERmFe9U4ZiA0M4kbT5nmuk3I85Sk4bagA==",
		"https://sabre.ct.comodo.com/",
	},
}

// addChain is a list of base64 encoded certificate chains, which will be submitted to the log server
type addChain struct {
	Chain []string `json:"chain"`
}

func submitCT(c *cli.Context) error {
	if !c.IsSet("input") {
		return errors.New("please specify input file")
	}

	if !c.IsSet("output") {
		return errors.New("please specify an output directory")
	}

	// check for optional file name flag
	fileName := ""

	if c.IsSet("filename") {
		fileName = c.String("filename")
	}

	inputFile, output := c.String("input"), c.String("output")

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return errors.New("please specify a valid input file")
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		return errors.New("please specify a valid output directory")
	}

	payload := getPayload(inputFile)

	var wg sync.WaitGroup
	wg.Add(len(ctLogList))

	// submit to all available log servers
	for _, e := range ctLogList {

		log.Printf("Submitting certificate to %s CT log", e.Name)

		if fileName != "" {
			go e.Submit(payload, output+e.Name+"."+fileName+".sct", &wg)
		} else {
			go e.Submit(payload, output+e.Name+".sct", &wg)
		}
	}

	// wait for all goroutines to finish
	wg.Wait()
	log.Println("Successfully submitted CT logs to all available servers!")
	log.Println("Do not forget to include the output directory in your NginX host config using 'ssl_ct_static_scts'")

	return nil
}

// getPayload parses the given pem file into the addChain struct and returns it's binary representation
func getPayload(input string) []byte {
	f, err := ioutil.ReadFile(input)

	if err != nil {
		log.Fatalf("Cannt find input file %s!", input)
		os.Exit(1)
	}

	// parsing given certificate
	msg := addChain{}
	for {
		block, remaining := pem.Decode(f)
		f = remaining

		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		msg.Chain = append(msg.Chain, base64.StdEncoding.EncodeToString(block.Bytes))
	}

	// parsing of given file wasn't possible, so termiante
	if len(msg.Chain) == 0 {
		log.Fatalf("Failed parsing given PEM certificate. Please check format!")
		os.Exit(1)
	}

	// construct add-chain message
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("Failed contructing add-chain message! Error: %s", err)
		os.Exit(1)
	}

	return payload
}
