param()

Write-Host "Building and running deploy stack (collector + jaeger + gateway)"
Push-Location deploy
Start-Process -NoNewWindow -Wait -FilePath docker-compose -ArgumentList "up --build"
Pop-Location

Write-Host "Waiting 3s for services"
Start-Sleep -Seconds 3

Write-Host "Running demo script against gateway"
& ..\scripts\run.ps1 -UseDocker
