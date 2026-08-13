package health

import "testing"

func TestAggregateKeepsIndependentHealthDomains(t *testing.T){
	status,err:=Aggregate([]Check{{Domain:System,Status:Healthy},{Domain:Recovery,Status:Degraded,Reason:"restore drill overdue"}}); if err!=nil{t.Fatal(err)}
	if status!=Degraded{t.Fatalf("got %s",status)}
}
