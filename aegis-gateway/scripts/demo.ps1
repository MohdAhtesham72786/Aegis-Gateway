# Demo scripts for Aegis Gateway

# Blocked high-value payment
Write-Host "Blocked high-value payment (should 403)"
curl -s -H "Content-Type: application/json" -H "X-Agent-ID: finance-agent" -X POST http://localhost:8080/tools/payments/create -d '{"amount":50000,"currency":"USD","vendor_id":"V99"}' | jq .

# Allowed payment
Write-Host "Allowed payment (should 200)"
curl -s -H "Content-Type: application/json" -H "X-Agent-ID: finance-agent" -X POST http://localhost:8080/tools/payments/create -d '{"amount":100.5,"currency":"USD","vendor_id":"V1"}' | jq .

# Allowed HR file read
Write-Host "Allowed HR read (should 200)"
curl -s -H "Content-Type: application/json" -H "X-Agent-ID: hr-agent" -X POST http://localhost:8080/tools/files/read -d '{"path":"/hr-docs/employee1.txt"}' | jq .

# Blocked HR file read outside
Write-Host "Blocked HR read outside (should 403)"
curl -s -H "Content-Type: application/json" -H "X-Agent-ID: hr-agent" -X POST http://localhost:8080/tools/files/read -d '{"path":"/legal/contract.docx"}' | jq .
