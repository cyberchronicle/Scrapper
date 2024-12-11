push:
	docker buildx build \
         --platform linux/amd64,linux/arm64 \
         -t drc0de/scrapping:latest \
         --push .