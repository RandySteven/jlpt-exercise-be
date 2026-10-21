include ./files/env/.env
export

cmd_folder = ./cmd/
gorun = @go run

# ── Environment ────────────────────────────────────────────────
ifeq ($(ENV),prod)
    yaml_file = ./files/yaml/app.prod.yml
else ifeq ($(ENV),staging)
    yaml_file = ./files/yaml/app.docker.yml
else ifeq ($(ENV),dev)
    yaml_file = ./files/yaml/app.local.yml
else
    yaml_file = ./files/yaml/app.local.yml
endif

# ── Server ─────────────────────────────────────────────────────
.PHONY: run migration seed drop consumer refresh

run:
	$(gorun) $(cmd_folder)main -config $(yaml_file)

migration:
	$(gorun) $(cmd_folder)migration -config $(yaml_file)

seed:
	$(gorun) $(cmd_folder)seed -config $(yaml_file)

drop:
	$(gorun) $(cmd_folder)drop -config $(yaml_file)

consumer:
	$(gorun) $(cmd_folder)consumers -config $(yaml_file)

refresh: drop migration seed

# ── Docker ─────────────────────────────────────────────────────
.PHONY: run-docker stop-docker

run-docker:
	docker compose up --build -d

stop-docker:
	docker compose down

# ── Codegen ────────────────────────────────────────────────────
.PHONY: make_model make_workflow gen_proto

make_model:
	@read -p "Model name: " MODELNAME; \
	MODELFILE=$$(echo $$MODELNAME | sed 's/^./\l&/'); \
	echo "Creating model: $$MODELFILE.go"; \
	echo "package model" > $$MODELFILE.go; \
	echo "type $$MODELNAME struct {}" >> $$MODELFILE.go; \
	echo "Creating repository: $${MODELFILE}_repository.go"; \
	printf 'package repositories\n\nimport "models"\n\ntype I%sRepository interface {\n    IRepository[models.%s]\n}\n' \
		$$MODELNAME $$MODELNAME > $${MODELFILE}_repository.go

make_workflow:
	@read -p "Workflow name: " WORKFLOWNAME; \
	WORKFLOWNAME=$$(echo "$$WORKFLOWNAME" | tr -d '[:space:]'); \
	test -n "$$WORKFLOWNAME" || { echo "Workflow name is required"; exit 1; }; \
	DIR="logic/$${WORKFLOWNAME}_workflow"; \
	PKG="$${WORKFLOWNAME}_workflow"; \
	STRUCT="$${WORKFLOWNAME}Workflow"; \
	FIRST=$$(printf '%s' "$$WORKFLOWNAME" | cut -c1 | tr '[:lower:]' '[:upper:]'); \
	REST=$$(printf '%s' "$$WORKFLOWNAME" | cut -c2-); \
	INTERFACE="$${FIRST}$${REST}Workflow"; \
	WF_FILE="$$DIR/workflow.go"; \
	EXEC_FILE="$$DIR/execution_data.go"; \
	mkdir -p "$$DIR"; \
	echo "Creating workflow at $$WF_FILE"; \
	printf '%s\n' \
		"package $$PKG" \
		"" \
		"import (" \
		"	temporal_client \"github.com/RandySteven/go-cook/temporal\"" \
		")" \
		"" \
		"type (" \
		"	$$STRUCT struct {" \
		"		workflow       temporal_client.WorkflowExecution" \
		"		temporalClient temporal_client.Temporal" \
		"	}" \
		"" \
		"	$$INTERFACE interface {}" \
		")" \
		> "$$WF_FILE"; \
	echo "Creating execution data at $$EXEC_FILE"; \
	printf '%s\n' \
		"package $$PKG" \
		"" \
		"import (" \
		"	\"encoding/json\"" \
		"" \
		"	temporal_client \"github.com/RandySteven/go-cook/temporal\"" \
		")" \
		"" \
		"type ExecutionData struct {" \
		"	WorkflowID   string \`json:\"workflow_id\"\`" \
		"	RunID        string \`json:\"run_id\"\`" \
		"	CurrActivity string \`json:\"curr_activity\"\`" \
		"	Status       string \`json:\"status\"\`" \
		"}" \
		"" \
		"var _ temporal_client.ExecutionData = (*ExecutionData)(nil)" \
		"" \
		"func (e *ExecutionData) Clone() temporal_client.ExecutionData {" \
		"	c := *e" \
		"	return &c" \
		"}" \
		"" \
		"func (e *ExecutionData) GetActivity() string {" \
		"	return e.CurrActivity" \
		"}" \
		"" \
		"func (e *ExecutionData) Marshal() ([]byte, error) {" \
		"	return json.Marshal(e)" \
		"}" \
		"" \
		"func (e *ExecutionData) SetActivity(activityName string) {" \
		"	e.CurrActivity = activityName" \
		"}" \
		"" \
		"func (e *ExecutionData) Unmarshal(data []byte) error {" \
		"	return json.Unmarshal(data, e)" \
		"}" \
		> "$$EXEC_FILE"

gen_proto:
	protoc \
		--go_out=./proto \
		--go-grpc_out=./proto \
		--proto_path=./proto \
		./proto/service.proto

# ── Test ───────────────────────────────────────────────────────
.PHONY: test test-cover

test:
	go test -count=1 ./...

test-cover:
	go test -count=1 -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report: coverage.html"

# ── Debug ──────────────────────────────────────────────────────
.PHONY: env_check

env_check:
	@echo "ENV:       $(ENV)"
	@echo "yaml_file: $(yaml_file)"