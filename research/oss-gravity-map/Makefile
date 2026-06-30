# OSS Gravity Map — Makefile
#
# Empirical validation of EIS against 25 major open-source projects.
# Detects architects, hidden architects, entropy fighters, and collapse risk.

SCRIPTS := scripts
DATA := data
ANALYSIS := analysis

.PHONY: all clone analyze ground-truth report ingest clean

# Full pipeline
all: clone ground-truth analyze report

# Step 1: Clone all 25 repositories (requires ~50GB disk space)
clone:
	@echo "=== Step 1: Cloning repositories ==="
	@bash $(SCRIPTS)/clone-repos.sh

# Step 2: Fetch ground truth from GitHub API (requires gh CLI)
ground-truth:
	@echo "=== Step 2: Fetching ground truth ==="
	@bash $(SCRIPTS)/fetch-ground-truth.sh

# Step 3: Run EIS analysis on all repos (requires eis CLI)
analyze:
	@echo "=== Step 3: Running EIS analysis ==="
	@bash $(SCRIPTS)/analyze-repos.sh

# Step 4: Generate analysis report
report:
	@echo "=== Step 4: Generating report ==="
	@cd "$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))" && \
		python3 $(SCRIPTS)/analyze-results.py \
			--results-dir $(DATA)/results \
			--ground-truth-dir $(DATA)/ground-truth \
			--dataset dataset.yaml \
			--output-dir $(ANALYSIS)

# Analyze a single repo (usage: make single REPO=react)
single:
	@echo "Analyzing $(REPO)..."
	@eis analyze --format json $(DATA)/repos/$(REPO) > $(DATA)/results/$(REPO).json
	@echo "Done. Results: $(DATA)/results/$(REPO).json"

# Publish an EIS research run to the OrbitLens research store (one repo at a
# time, disk-bounded). Reads ingest-repos.yaml; resolves git authors to GitHub
# ids; POSTs to $API_BASE (default stg). Env: RESEARCH_INGEST_TOKEN (required),
# GITHUB_TOKEN (recommended), API_BASE, RUN_ID, ONLY, DRY_RUN. See run-ingest.sh.
ingest:
	@bash $(SCRIPTS)/run-ingest.sh

# Clean generated data (keeps cloned repos)
clean:
	rm -rf $(DATA)/results/*.json
	rm -rf $(DATA)/ground-truth/*.json
	rm -rf $(ANALYSIS)/*

# Clean everything including cloned repos
clean-all: clean
	rm -rf $(DATA)/repos/*
