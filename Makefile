.PHONY: prepare-test
prepare-test:
	docker build -t warashi/nginx:ok testdata/healthcheck/success
	docker build -t warashi/nginx:ng testdata/healthcheck/fail
	docker build -t warashi/nginx:none testdata/healthcheck/none

.PHONY: test
test: prepare-test
	go test ./...
