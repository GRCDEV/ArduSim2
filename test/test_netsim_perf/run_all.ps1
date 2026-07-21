$flushInterval = 10
$logFile = "run_all_results.log"
Clear-Content $logFile -ErrorAction SilentlyContinue

Write-Host "Running with flush_interval_ms = $flushInterval"
Add-Content $logFile "========== FLUSH INTERVAL = $flushInterval ms =========="

docker-compose up --build --abort-on-container-exit > "compose_$flushInterval.log" 2>&1

# Extract results: lines from tester with numeric content or result headers
$results = Select-String -Path "compose_$flushInterval.log" -Pattern "tester-1\s+\|" | Select-Object -ExpandProperty Line
Add-Content $logFile $results

Write-Host "All done! Results saved to $logFile"
Write-Host "Full compose log: compose_$flushInterval.log"
