package health

import "testing"

func completeChecks() []Check {
	return []Check{
		{Domain:System,Status:Healthy},
		{Domain:Knowledge,Status:Healthy},
		{Domain:Authority,Status:Healthy},
		{Domain:Source,Status:Healthy},
		{Domain:Recovery,Status:Healthy},
	}
}

func TestAggregateKeepsIndependentHealthDomains(t *testing.T){
	checks:=completeChecks()
	checks[4]=Check{Domain:Recovery,Status:Degraded,Reason:"restore drill overdue"}
	status,err:=Aggregate(checks); if err!=nil{t.Fatal(err)}
	if status!=Degraded{t.Fatalf("got %s",status)}
}

func TestAggregateRejectsMissingDomain(t *testing.T){
	checks:=completeChecks()[:4]
	if _,err:=Aggregate(checks);err==nil{t.Fatal("expected missing-domain failure")}
}

func TestAggregateRejectsDuplicateDomain(t *testing.T){
	checks:=completeChecks()
	checks=append(checks,Check{Domain:System,Status:Healthy})
	if _,err:=Aggregate(checks);err==nil{t.Fatal("expected duplicate-domain failure")}
}
