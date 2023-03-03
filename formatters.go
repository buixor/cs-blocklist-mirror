package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/crowdsecurity/crowdsec/pkg/models"
)

var FormattersByName map[string]func(w http.ResponseWriter, r *http.Request) = map[string]func(w http.ResponseWriter, r *http.Request){
	"plain_text": PlainTextFormatter,
	"microtik":   MicroTikFormatter,
}

func PlainTextFormatter(w http.ResponseWriter, r *http.Request) {
	decisions := r.Context().Value(globalDecisionRegistry.Key).([]*models.Decision)
	ips := make([]string, len(decisions))
	for i, decision := range decisions {
		ips[i] = *decision.Value
	}
	sort.Strings(ips)
	w.Write([]byte(strings.Join(ips, "\n")))
}

func MicroTikFormatter(w http.ResponseWriter, r *http.Request) {
	decisions := r.Context().Value(globalDecisionRegistry.Key).([]*models.Decision)
	ips := make([]string, len(decisions))
	listName := r.URL.Query().Get("listname")
	if listName == "" {
		listName = "CrowdSec"
	}
	for i, decision := range decisions {
		var ipType = "/ip"
		if strings.Contains(*decision.Value, ":") {
			ipType = "/ipv6"
		}
		ips[i] = fmt.Sprintf(
			"%s firewall address-list add list=%s address=%s comment=\"%s for %s\"",
			ipType,
			listName,
			*decision.Value,
			*decision.Scenario,
			*decision.Duration,
		)
	}
	sort.Strings(ips)
	if !r.URL.Query().Has("ipv6only") {
		w.Write([]byte(fmt.Sprintf("/ip firewall address-list remove [find list=%s]\n", listName)))
	}
	if !r.URL.Query().Has("ipv4only") {
		w.Write([]byte(fmt.Sprintf("/ipv6 firewall address-list remove [find list=%s]\n", listName)))
	}
	w.Write([]byte(strings.Join(ips, "\n")))
}
