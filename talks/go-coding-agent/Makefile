# This example creates simple agent using gpt-oss model  
example01-step1:
	go run cmd/examples/example-01/step-1/main.go

example01-step2:
	go run cmd/examples/example-01/step-2/main.go

example01-step3:
	go run cmd/examples/example-01/step-3/main.go

example01-step4:
	go run cmd/examples/example-01/step-4/main.go

kronk-up:
	export KRONK_CACHE_MODEL_CONFIG_FILE=zarf/kms/model_config.yaml && \
	kronk server start

kronk-logs:
	kronk server logs

kronk-list-models:
	kronk model list --local