$tag = Get-Date -Format "yyyy-MM-dd-HH-mm"
$imageName = "harbor.andrusdiaz.dev/pabilo/pabilo-gobcv:$tag"

Write-Host "Building Docker image: $imageName"
docker build -t $imageName .

if ($LASTEXITCODE -eq 0) {
    Write-Host "Pushing Docker image: $imageName"
    docker push $imageName
} else {
    Write-Error "Docker build failed."
    exit 1
}
