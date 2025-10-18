param(
    [switch]$UseDocker
)

if ($UseDocker) {
    Write-Host "Building Docker image..."
    Start-Process -NoNewWindow -Wait -FilePath docker -ArgumentList "build -t aegis:local .."
    Write-Host "Running aegis container..."
    Start-Process -NoNewWindow -FilePath docker -ArgumentList "run --rm -p 8080:8080 aegis:local"
    Start-Sleep -Seconds 2
} else {
    Write-Host "Running with local Go..."
    Push-Location ..
    go run ./cmd/aegis
    Pop-Location
}

function PostJson($url, $json, $agent) {
    $headers = @{ 'X-Agent-ID' = $agent; 'Content-Type' = 'application/json' }
    try {
        $resp = Invoke-RestMethod -Method Post -Uri $url -Body $json -Headers $headers
        $s = $resp | ConvertTo-Json -Depth 5
        Write-Host $s
    } catch {
        Write-Host "Request failed: $_"
    }
}

Write-Host "Run demo (blocked high-value payment):"
PostJson "http://localhost:8080/tools/payments/create" '{"amount":50000,"currency":"USD","vendor_id":"V99"}' "finance-agent"

Write-Host "Allowed payment:"
PostJson "http://localhost:8080/tools/payments/create" '{"amount":100.5,"currency":"USD","vendor_id":"V1"}' "finance-agent"

Write-Host "Allowed HR read:"
PostJson "http://localhost:8080/tools/files/read" '{"path":"/hr-docs/employee1.txt"}' "hr-agent"

Write-Host "Blocked HR read outside:"
PostJson "http://localhost:8080/tools/files/read" '{"path":"/legal/contract.docx"}' "hr-agent"
