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
.PHONY: make_model gen_proto

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
	DIR="logic/$${WORKFLOWNAME}_workflow"; \
	mkdir -p $$DIR; \
	PKG="$${WORKFLOWNAME}_workflow"; \
	STRUCT="$${WORKFLOWNAME}Workflow"; \
	INTERFACE="$$(echo $$WORKFLOWNAME | sed 's/^./\u&/')Workflow"; \
	FILE="$$DIR/workflow.go"; \
	echo "Creating workflow at $$FILE"; \
	printf 'package %s\n\nimport "github.com/RandySteven/go-cook/temporal_client"\n\ntype %s struct {}\n\ntype %s interface {\n    IWorkflow[temporal_client.%s]\n}\n' \
		"$$PKG" "$$STRUCT" "$$INTERFACE" "$$(echo $$WORKFLOWNAME | sed 's/^./\u&/')" > "$$FILE"

gen_proto:
	protoc \
		--go_out=./proto \
		--go-grpc_out=./proto \
		--proto_path=./proto \
		./proto/service.proto

# ── Debug ──────────────────────────────────────────────────────
.PHONY: env_check

env_check:
	@echo "ENV:       $(ENV)"
	@echo "yaml_file: $(yaml_file)"