run-local:
	go run .

vuln-check:
	govulncheck .

deploy:
	caprover deploy --default

run-prod-locally:
	docker build -t email-linker .
	docker run -p 8080:8080 --env-file .env --env GIN_MODE=release email-linker

e2e-test:
	go test . -v