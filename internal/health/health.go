package health

import "fmt"

type Domain string
type Status string

const (
	System    Domain = "system"
	Knowledge Domain = "knowledge"
	Authority Domain = "authority"
	Source    Domain = "source"
	Recovery  Domain = "recovery"

	Healthy  Status = "healthy"
	Unknown  Status = "unknown"
	Degraded Status = "degraded"
	Blocked  Status = "blocked"
)

var requiredDomains = []Domain{System, Knowledge, Authority, Source, Recovery}

type Check struct {
	Domain Domain `json:"domain"`
	Status Status `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func (c Check) Validate() error {
	switch c.Domain { case System,Knowledge,Authority,Source,Recovery: default: return fmt.Errorf("HEALTH_INVALID: unknown domain %q",c.Domain) }
	switch c.Status { case Healthy,Unknown,Degraded,Blocked: default: return fmt.Errorf("HEALTH_INVALID: unknown status %q",c.Status) }
	if c.Status!=Healthy && c.Reason=="" { return fmt.Errorf("HEALTH_INVALID: non-healthy status requires reason") }
	return nil
}

func Aggregate(checks []Check) (Status,error) {
	rank:=map[Status]int{Healthy:0,Unknown:1,Degraded:2,Blocked:3}
	seen:=map[Domain]bool{}
	worst:=Healthy
	for _,c:=range checks {
		if err:=c.Validate(); err!=nil{return "",err}
		if seen[c.Domain] { return "",fmt.Errorf("HEALTH_INVALID: duplicate domain %q",c.Domain) }
		seen[c.Domain]=true
		if rank[c.Status]>rank[worst]{worst=c.Status}
	}
	for _,domain:=range requiredDomains {
		if !seen[domain] { return "",fmt.Errorf("HEALTH_INCOMPLETE: missing domain %q",domain) }
	}
	return worst,nil
}
