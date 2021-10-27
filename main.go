package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

// http://sean.lyn.ch/2015/02/parsing-json-in-go/
type TicketData struct {
	Data struct {
		Ticket string `json:"ticket"`
	} `json:"data"`
}

type Node struct {
	Name string `json:"node"`
}

type NodesData struct {
	Data []Node `json:"data"`
}

type Lxc struct {
	Id   string `json:"vmid"`
	Name string `json:"name"`
}

type LxcsData struct {
	Data []Lxc `json:"data"`
}

type Virtual struct {
	Id   string `json:"vmid"`
	Name string `json:"name"`
}

type VirtualsData struct {
	Data []Virtual `json:"data"`
}

type Rrrddata struct {
	Mem float64 `json:"mem"`
	Cpu float64 `json:"cpu"`
}

type RrrddataData struct {
	Data []Rrrddata `json:"data"`
}

var proxmox_api_url string = os.Getenv("PROXMOX_HOST") + "/api2/json"

var client http.Client

func init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}
	client = http.Client{
		Jar: jar,
	}
}

func auth() string {

	data := url.Values{
		"username": {os.Getenv("PROXMOX_USER") + "@pam"},
		"password": {os.Getenv("PROXMOX_PASSWORD")},
	}

	resp, err := http.PostForm(proxmox_api_url+"/access/ticket", data)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var res TicketData

	json.NewDecoder(resp.Body).Decode(&res)

	return res.Data.Ticket

}

func request(auth string, resource_path string) http.Response {

	cookie := &http.Cookie{
		Name:   "PVEAuthCookie",
		Value:  auth,
		MaxAge: 300,
	}

	req, err := http.NewRequest("GET", proxmox_api_url+resource_path, nil)
	if err != nil {
		panic(nil)
	}
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	if err != nil {
		panic(nil)
	}
	return *resp
}

func nodes(auth string) []Node {
	response := request(auth, "/nodes/")

	var nodes NodesData

	json.NewDecoder(response.Body).Decode(&nodes)

	return nodes.Data
}

func virtuals(auth string, virtualisation_type string, node_name string) []Virtual {
	response := request(auth, "/nodes/"+node_name+"/"+virtualisation_type)

	var nodes VirtualsData

	json.NewDecoder(response.Body).Decode(&nodes)

	return nodes.Data
}

func rrddatas(auth string, virtualisation_type string, node_name string, lxc_id string) []Rrrddata {

	response := request(auth, "/nodes/"+node_name+"/"+virtualisation_type+"/"+lxc_id+"/rrddata?timeframe=month")

	var rrddatadata RrrddataData

	json.NewDecoder(response.Body).Decode(&rrddatadata)

	return rrddatadata.Data
}

func main() {

	auth := auth()

	agregator_mem := make(map[string]float64)
	agregator_cpu := make(map[string]float64)

	for _, node := range nodes(auth) {

		for _, virtualsation_type := range [2]string{"lxc", "qemu"} {

			for _, lxc := range virtuals(auth, virtualsation_type, node.Name) {

				var mem_nonzero_counter float64
				var cpu_nonzero_counter float64

				var mem float64
				var cpu float64

				for _, rrddata := range rrddatas(auth, virtualsation_type, node.Name, lxc.Id) {

					cpu += float64(rrddata.Cpu * 100)
					cpu_nonzero_counter += 1

					if rrddata.Mem != 0 {
						mem += rrddata.Mem
						mem_nonzero_counter += 1
					}
				}

				cpu_avg := cpu / cpu_nonzero_counter
				var mem_avg float64
				if mem == 0 {
					mem_avg = 0
				} else {
					mem_avg = mem / mem_nonzero_counter / 1024 / 1024
				}

				name := strings.Split(lxc.Name, "-")[0]

				agregator_mem[name] += mem_avg
				agregator_cpu[name] += cpu_avg
			}

		}

	}

	for service_name := range agregator_mem {
		fmt.Printf("%s, %f, %f\n", service_name, agregator_cpu[service_name], agregator_mem[service_name])
	}

}
