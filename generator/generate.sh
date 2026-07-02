#!/usr/bin/env bash
# Assembles self-contained level branches from the main-branch sources.
#
#   generate.sh --out DIR SYSTEM LEVEL     assemble one level into DIR
#                                          (used by `make start` on main)
#   generate.sh --push [SYSTEM...]         assemble every level of the given
#                                          systems (default: all) and force-
#                                          push level-N-*/SYSTEM branches
#
# Flow is strictly one-directional: main → generated branches, never back.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

level_dir()  { case "$1" in
  1) echo "level-1-observe";; 2) echo "level-2-experiment";;
  3) echo "level-3-build";;   4) echo "level-4-fix";;
  5) echo "level-5-scratch";; *) echo "unknown"; return 1;; esac; }
level_name() { case "$1" in
  1) echo "Observe & Understand";; 2) echo "Tweak & Experiment";;
  3) echo "Build the Missing Piece";; 4) echo "Fix the Broken System";;
  5) echo "Build from Scratch";; esac; }
level_time() { case "$1" in
  1) echo "~1 hour";; 2) echo "2-3 hours";; 3) echo "2-4 hours";;
  4) echo "2-4 hours";; 5) echo "6-12 hours";; esac; }

src_sha() { git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo "unknown"; }

assemble() {
  local system="$1" level="$2" out="$3"
  local sys_dir="$ROOT/systems/$system"
  local lvl_dir="$sys_dir/$(level_dir "$level")"
  local shared="$sys_dir/shared"
  local sha; sha="$(src_sha)"

  [[ -d "$lvl_dir" ]] || { echo "error: $lvl_dir does not exist" >&2; return 1; }

  rm -rf "$out"
  mkdir -p "$out/system"

  # 1. Infrastructure and tooling — copied verbatim, identical on every branch.
  cp -r "$ROOT/infrastructure" "$out/infrastructure"
  cp -r "$ROOT/chaos-toolkit"  "$out/chaos-toolkit"
  cp -r "$ROOT/cli"            "$out/cli"
  mkdir -p "$out/docs"
  cp -r "$ROOT/docs/ai-failure-cases" "$out/docs/ai-failure-cases"
  cp -r "$ROOT/.devcontainer" "$out/.devcontainer"
  cp "$ROOT/Makefile" "$ROOT/.gitattributes" "$ROOT/.gitignore" "$ROOT/.env.example" "$out/"

  # 2. system/ — the only thing that changes between levels.
  cp -r "$shared/services"   "$out/system/services"
  cp -r "$shared/load-tests" "$out/system/load-tests"
  cp -r "$shared/db"         "$out/system/db"
  cp "$shared/README.md"     "$out/system/README.md"
  cp "$shared/docker-compose.yml" "$out/system/docker-compose.yml"
  cp "$sys_dir/JOURNAL_TEMPLATE.md" "$out/system/JOURNAL_TEMPLATE.md"

  # sed_escape: '&' and '\' are special in sed replacements ("Observe &
  # Understand" would otherwise render as the matched placeholder).
  sed_escape() { printf '%s' "$1" | sed 's/[&\\/]/\\&/g'; }
  sed -e "s/__LEVEL__/$level/" \
      -e "s/__LEVEL_NAME__/$(sed_escape "$(level_name "$level")")/" \
      -e "s/__ESTIMATED_TIME__/$(sed_escape "$(level_time "$level")")/" \
      -e "s/__GENERATED_FROM__/main@$sha/" \
      "$shared/system.yaml.tmpl" > "$out/system/system.yaml"

  # 3. Level overlay: docs into system/, config to the branch root.
  find "$lvl_dir" -maxdepth 1 -name '*.md' -exec cp {} "$out/system/" \;
  cp "$lvl_dir/config.yaml" "$out/config.yaml"
  cp "$lvl_dir/config.yaml" "$out/.config.baseline.yaml"

  # 4. Level-specific surgery.
  case "$level" in
    3)
      cp "$lvl_dir/stubs/ring.go" "$out/system/services/internal/ring/ring.go"
      ;;
    4)
      cp -r "$lvl_dir/.solutions" "$out/system/.solutions"
      ;;
    5)
      for svc in api-gateway shortener redirector; do
        cp "$lvl_dir/stubs/$svc/main.go" "$out/system/services/$svc/main.go"
      done
      # The ring is part of the scratch build too — reuse the L3 stub.
      cp "$sys_dir/level-3-build/stubs/ring.go" "$out/system/services/internal/ring/ring.go"
      cp -r "$lvl_dir/contracts" "$out/system/contracts"
      ;;
  esac

  echo "main@$sha" > "$out/.generated-from"
  echo "assembled $system level $level -> $out"
}

push_branches() {
  local systems=("$@")
  [[ ${#systems[@]} -gt 0 ]] || systems=(url-shortener)
  local origin; origin="$(git -C "$ROOT" remote get-url origin)"
  local sha; sha="$(src_sha)"
  local work; work="$(mktemp -d)"
  trap 'rm -rf "$work"' EXIT

  for system in "${systems[@]}"; do
    for level in 1 2 3 4 5; do
      local branch="$(level_dir "$level")/$system"
      local out="$work/$branch"
      assemble "$system" "$level" "$out"
      git -C "$out" init -q -b "$branch"
      git -C "$out" add -A
      git -C "$out" -c user.name="systemdesignlab-bot" \
                    -c user.email="bot@systemdesignlab" \
                    commit -qm "Generated from main@$sha"
      git -C "$out" push --force "$origin" "$branch:refs/heads/$branch"
      echo "pushed $branch"
    done
  done
}

case "${1:-}" in
  --out)
    [[ $# -eq 4 ]] || { echo "usage: generate.sh --out DIR SYSTEM LEVEL" >&2; exit 2; }
    assemble "$3" "$4" "$2"
    ;;
  --push)
    shift
    push_branches "$@"
    ;;
  *)
    echo "usage: generate.sh --out DIR SYSTEM LEVEL | --push [SYSTEM...]" >&2
    exit 2
    ;;
esac
