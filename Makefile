# SystemDesignLab
#
# Two modes, one Makefile:
#  - On a generated level branch (system/ exists): targets act directly.
#  - On main (source of truth): targets assemble the requested SYSTEM/LEVEL
#    into .lab/ via the same generator CI uses, then delegate. This means
#    every `make start` on main is also a test of branch generation.
#
#    make start SYSTEM=url-shortener LEVEL=1

SYSTEM   ?= url-shortener
LEVEL    ?= 1
SCENARIO ?= steady-state

.PHONY: help start stop redeploy clean load-test dashboard validate journal \
        reveal-solution diagnose switch-cache generate-branches \
        chaos-kill-cache chaos-lag-database chaos-overload chaos-restore \
        start-parallel clean-parallel test

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# The CLI is built on demand; users need Go (a listed prerequisite).
bin/sdl: cli/*.go cli/go.mod cli/diagnose/*.yaml
	@mkdir -p bin && cd cli && go build -o ../bin/sdl .

diagnose: bin/sdl ## Diagnostic quiz → suggests your entry level
	@./bin/sdl diagnose

level-1 level-2 level-3 level-4 level-5: bin/sdl ## Hop levels; parks/restores uncommitted work
	@./bin/sdl switch $(subst level-,,$@)

ifneq ($(wildcard system/docker-compose.yml),)
# ============================ BRANCH MODE ====================================
CACHE_PROVIDER := $(shell awk '/^cache:/{f=1} f && /provider:/{print $$2; exit}' config.yaml)
COMPOSE := COMPOSE_PROFILES=$(CACHE_PROVIDER) docker compose -f system/docker-compose.yml
# For teardown, cover both profiles — after a provider switch the OLD
# profile's containers must go too (they hold the cache-1..3 aliases).
COMPOSE_ALL := COMPOSE_PROFILES=redis,memcached docker compose -f system/docker-compose.yml

# Guardrail: generated level-* branches are force-pushed by CI, so work
# committed there gets stranded. On the code levels (3-5) we auto-switch to
# a personal my-progress branch; on 1-2 (nothing to commit, and peeking at
# level-1 for a healthy refresher is a supported flow) we only notify.
# Opt out with GUARD=off.
.PHONY: .guard-branch
.guard-branch:
	@if [ "$(GUARD)" = "off" ]; then exit 0; fi; \
	branch=$$(git symbolic-ref --short HEAD 2>/dev/null || echo ""); \
	case "$$branch" in level-*) \
	  sys=$$(awk '/^name:/{print $$2; exit}' system/system.yaml); \
	  lvl=$$(awk '/^level:/{print $$2; exit}' system/system.yaml); \
	  mine="my-progress/$$sys-level-$$lvl"; \
	  if [ "$$lvl" -ge 3 ]; then \
	    if git rev-parse --verify -q "$$mine" >/dev/null; then \
	      echo "==> '$$branch' is generated (CI force-pushes it). Switching to your '$$mine'."; \
	      git checkout -q "$$mine"; \
	    else \
	      echo "==> '$$branch' is generated (CI force-pushes it) — commits here get stranded."; \
	      echo "==> Created '$$mine' for your work and switched to it."; \
	      git checkout -q -b "$$mine"; \
	    fi; \
	  else \
	    echo "note: '$$branch' is a generated branch. Fine for observing/experimenting;"; \
	    echo "      don't commit to it (use: git checkout -b $$mine). GUARD=off silences this."; \
	  fi;; \
	esac

start: .guard-branch ## Build and start the stack
	$(COMPOSE) up -d --build
	@echo
	@echo "  Gateway     http://localhost:8080"
	@echo "  Grafana     http://localhost:3000   (make dashboard)"
	@echo "  Prometheus  http://localhost:9090"
	@echo
	@echo "  Next: make load-test   then read system/CONTEXT.md"

stop: ## Stop containers (keep volumes)
	$(COMPOSE_ALL) stop

redeploy: .guard-branch ## Apply config.yaml changes: down + up (volumes preserved)
	$(COMPOSE_ALL) down --remove-orphans
	$(COMPOSE) up -d --build

clean: ## Tear down containers AND volumes
	$(COMPOSE_ALL) down -v --remove-orphans

load-test: ## Run k6 (SCENARIO=steady-state|read-spike|hot-key, RATE=, DURATION=)
	$(COMPOSE) run --rm -e RATE="$(RATE)" -e DURATION="$(DURATION)" k6 run /scripts/$(SCENARIO).js

dashboard: ## Open Grafana
	@xdg-open http://localhost:3000 2>/dev/null || open http://localhost:3000 2>/dev/null \
	  || echo "Open http://localhost:3000 in your browser"

validate: bin/sdl ## Run level-appropriate checks
	@./bin/sdl validate

journal: bin/sdl ## Create my-journal.md from the template
	@./bin/sdl journal

reveal-solution: bin/sdl ## Unhide SOLUTIONS.md (Level 4)
	@./bin/sdl reveal-solution

switch-cache: bin/sdl ## Swap cache provider: make switch-cache PROVIDER=memcached
	@test -n "$(PROVIDER)" || { echo "usage: make switch-cache PROVIDER=redis|memcached"; exit 2; }
	@./bin/sdl switch-cache $(PROVIDER)

chaos-kill-cache: ## Kill all cache nodes for 60s (OUTAGE= to change)
	@OUTAGE="$(OUTAGE)" bash chaos-toolkit/scripts/kill-cache.sh

chaos-lag-database: ## Add 300ms latency to the DB (DELAY= to change)
	@DELAY="$(DELAY)" bash chaos-toolkit/scripts/lag-database.sh

chaos-overload: ## 5x traffic for 2 minutes (RATE=, DURATION=)
	@RATE="$(RATE)" DURATION="$(DURATION)" bash chaos-toolkit/scripts/overload.sh

chaos-restore: ## Undo all chaos
	@bash chaos-toolkit/scripts/restore.sh

test: ## Run the system's Go unit tests
	cd system/services && go test ./internal/...

else
# ============================= MAIN MODE =====================================
LAB := .lab/$(SYSTEM)-level-$(LEVEL)

DELEGATED := start stop redeploy clean load-test dashboard validate journal \
             reveal-solution switch-cache chaos-kill-cache chaos-lag-database \
             chaos-overload chaos-restore

$(DELEGATED): .assemble ## (main mode: assembles SYSTEM/LEVEL into .lab/ first)
	@$(MAKE) --no-print-directory -C $(LAB) $@ SCENARIO=$(SCENARIO) PROVIDER=$(PROVIDER)

# Reassemble from sources on every use (so source edits propagate), but
# keep live state that isn't source: a user-modified config.yaml and the
# machine's calibrated .baseline.json.
.PHONY: .assemble
.assemble:
	@mkdir -p .lab
	@if [ -f $(LAB)/.baseline.json ]; then cp $(LAB)/.baseline.json .lab/.baseline.keep; fi
	@if [ -f $(LAB)/config.yaml ] && ! cmp -s $(LAB)/config.yaml $(LAB)/.config.baseline.yaml; then \
	  cp $(LAB)/config.yaml .lab/.config.keep; \
	  bash generator/generate.sh --out $(LAB) $(SYSTEM) $(LEVEL); \
	  mv .lab/.config.keep $(LAB)/config.yaml; \
	  echo "(kept your modified config.yaml)"; \
	else \
	  bash generator/generate.sh --out $(LAB) $(SYSTEM) $(LEVEL); \
	fi
	@if [ -f .lab/.baseline.keep ]; then mv .lab/.baseline.keep $(LAB)/.baseline.json; fi

generate-branches: ## Force-push all generated level branches (CI does this)
	bash generator/generate.sh --push

test: ## Run all Go tests (services + CLI build)
	cd systems/url-shortener/shared/services && go vet ./... && go test ./internal/...
	cd cli && go vet ./... && go build -o /dev/null .

endif

# Parallel worktrees (advanced): work on a second system without leaving
# your current branch. Requires the level branch to exist (generated by CI).
start-parallel: ## git worktree for a second system: make start-parallel SYSTEM=x LEVEL=1
	git worktree add .labs/$(SYSTEM) $(shell bash -c 'case $(LEVEL) in 1) echo level-1-observe;; 2) echo level-2-experiment;; 3) echo level-3-build;; 4) echo level-4-fix;; 5) echo level-5-scratch;; esac')/$(SYSTEM)
	@echo "Worktree ready: cd .labs/$(SYSTEM) && make start"

clean-parallel: ## Remove a parallel worktree
	git worktree remove .labs/$(SYSTEM)
