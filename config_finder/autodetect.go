package config_finder

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pivotal-cf-experimental/veritas/say"
)

func Autodetect(out io.Writer) error {
	jobs, err := ioutil.ReadDir("/var/vcap/jobs")
	if err != nil {
		return err
	}

	vitalsAddrs := []string{}
	executorAddr := ""
	gardenAddr := ""
	gardenNetwork := ""
	etcdCluster := ""
	receptorEndpoint := ""
	receptorUsername := ""
	receptorPassword := ""

	debugRe := regexp.MustCompile(`debugAddr=(\d+.\d+.\d+.\d+:\d+)`)
	etcdRe := regexp.MustCompile(`etcdCluster=\"(.+)\"`)
	executorRe := regexp.MustCompile(`listenAddr=(\d+.\d+.\d+.\d+:\d+)`)
	gardenTCPAddrRe := regexp.MustCompile(`gardenAddr=(\d+.\d+.\d+.\d+:\d+)`)
	gardenUnixAddrRe := regexp.MustCompile(`gardenAddr=([/\w+\.\d]+)`)
	receptorEndpointRe := regexp.MustCompile(`address=(\d+.\d+.\d+.\d+:\d+)`)
	receptorUsernameRe := regexp.MustCompile(`username=(\S*)\\`)
	receptorPasswordRe := regexp.MustCompile(`password=(\S*)\\`)

	for _, job := range jobs {
		jobDir := filepath.Join("/var/vcap/jobs", job.Name(), "bin")
		ctls, err := ioutil.ReadDir(jobDir)
		if err != nil {
			return err
		}

		for _, ctl := range ctls {
			if ctl.IsDir() {
				continue
			}
			if strings.HasSuffix(ctl.Name(), "_ctl") {
				name := strings.TrimSuffix(ctl.Name(), "_ctl")
				path := filepath.Join(jobDir, ctl.Name())
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				if debugRe.Match(data) {
					addr := string(debugRe.FindSubmatch(data)[1])
					vitalsAddrs = append(vitalsAddrs, fmt.Sprintf("%s:%s", name, addr))
				}

				if etcdRe.Match(data) {
					etcdCluster = string(etcdRe.FindSubmatch(data)[1])
					etcdCluster = strings.Replace(etcdCluster, `"`, ``, -1)
				}

				if name == "executor" && executorRe.Match(data) {
					executorAddr = "http://" + string(executorRe.FindSubmatch(data)[1])
				}

				if name == "executor" {
					if gardenTCPAddrRe.Match(data) {
						gardenAddr = string(gardenTCPAddrRe.FindSubmatch(data)[1])
						gardenNetwork = "tcp"
					} else if gardenUnixAddrRe.Match(data) {
						gardenAddr = string(gardenUnixAddrRe.FindSubmatch(data)[1])
						gardenNetwork = "unix"
					}
				}

				if name == "receptor" {
					if receptorEndpointRe.Match(data) {
						receptorEndpoint = string(receptorEndpointRe.FindSubmatch(data)[1])
					}
					if receptorUsernameRe.Match(data) {
						receptorUsername = string(receptorUsernameRe.FindSubmatch(data)[1])
					}
					if receptorPasswordRe.Match(data) {
						receptorPassword = string(receptorPasswordRe.FindSubmatch(data)[1])
					}
				}
			}
		}
	}

	if len(vitalsAddrs) > 0 {
		say.Fprintln(out, 0, "export VITALS_ADDRS=%s", strings.Join(vitalsAddrs, ","))
	}
	if executorAddr != "" {
		say.Fprintln(out, 0, "export EXECUTOR_ADDR=%s", executorAddr)
	}
	if gardenAddr != "" {
		say.Fprintln(out, 0, "export GARDEN_ADDR=%s", gardenAddr)
		say.Fprintln(out, 0, "export GARDEN_NETWORK=%s", gardenNetwork)
	}
	if etcdCluster != "" {
		say.Fprintln(out, 0, "export ETCD_CLUSTER=%s", etcdCluster)
	}
	if receptorEndpoint != "" {
		say.Fprintln(out, 0, "export RECEPTOR_ENDPOINT=%s", receptorEndpoint)
	}
	if receptorUsername != "" {
		say.Fprintln(out, 0, "export RECEPTOR_USERNAME=%s", receptorUsername)
	}
	if receptorPassword != "" {
		say.Fprintln(out, 0, "export RECEPTOR_PASSWORD=%s", receptorPassword)
	}

	return nil
}
