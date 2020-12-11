package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

func panicWithError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	machines := loadMachinesFromJSON()
	latencies := loadLatenciesFromJSON()
	assignCityToMachines(machines, latencies)

	for _, machine := range machines {
		fmt.Println(machine.city)
	}

	machineNameRuleListMap := produceRules(machines, latencies)

	for machineHostName, rules := range machineNameRuleListMap {
		fmt.Printf("=> %s\n", machineHostName)
		for _, rule := range rules {
			fmt.Printf("---------> %s\n", rule)
		}
	}

	produceBashFile(machineNameRuleListMap, machines)
}

func loadMachinesFromJSON() []Machine {

	var machines []Machine
	jsonFile, err := os.Open("./machines.json")
	panicWithError(err)

	byteValue, err := ioutil.ReadAll(jsonFile)
	panicWithError(err)

	err = json.Unmarshal(byteValue, &machines)
	panicWithError(err)

	return machines
}

func loadLatenciesFromJSON() []Latency {
	var latencies []Latency
	jsonFile, err := os.Open("./latencies.json")
	panicWithError(err)

	byteValue, err := ioutil.ReadAll(jsonFile)
	panicWithError(err)

	err = json.Unmarshal(byteValue, &latencies)
	panicWithError(err)

	return latencies
}

func assignCityToMachines(machines []Machine, latencies []Latency) {

	numLatency := len(latencies)
	for index := range machines {
		cityIndex := index % numLatency
		machines[index].city = latencies[cityIndex].From
	}

}

func produceRules(machines []Machine, latencies []Latency) map[string][]string {

	template := "sudo tcset --add eno1 --dst-network %s --delay %dms"

	machineNameRuleListMap := make(map[string][]string)
	for _, machine := range machines {
		machineNameRuleListMap[machine.HostName] = []string{}
		for _, otherMachine := range machines {

			if machine.city != otherMachine.city {

				city := findLatency(machine.city, latencies)
				machineNameRuleListMap[machine.HostName] = append(machineNameRuleListMap[machine.HostName], fmt.Sprintf(template, otherMachine.IPAddress, city.Values[otherMachine.city]))

			}

		}
	}

	return machineNameRuleListMap
}

func findLatency(from string, latencies []Latency) Latency {

	for _, latency := range latencies {
		if from == latency.From {
			return latency
		}
	}

	panic(fmt.Errorf("Could not find city %s in the latency list", from))
}

type rule struct {
	FromCity  string
	IPAddress string
	HostName  string
	Rules     []string
}

func produceBashFile(machineNameRuleListMap map[string][]string, machines []Machine) {

	file, err := os.Open("script-template.txt")
	panicWithError(err)

	templateBytes, err := ioutil.ReadAll(file)
	panicWithError(err)

	templateString := string(templateBytes)

	t := template.New("t")
	t, err = t.Parse(templateString)
	panicWithError(err)

	bashFile, err := os.Create("tc_rules.sh")
	panicWithError(err)

	for hostName, rules := range machineNameRuleListMap {

		machine := findMachine(hostName, machines)
		data := rule{FromCity: machine.city, IPAddress: machine.IPAddress, HostName: strings.Split(hostName, ".")[0], Rules: rules}
		err = t.Execute(bashFile, data)
		if err != nil {
			panic(err)
		}

		//return
	}

}

func findMachine(hostName string, machines []Machine) Machine {
	for _, machine := range machines {
		if hostName == machine.HostName {
			return machine
		}
	}

	panic(fmt.Errorf("Could not find machine with hostname %s in the machine list", hostName))
}
