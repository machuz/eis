#!/usr/bin/env python3
"""
Generate per-repository eis.yaml config files.

Architecture patterns are language/framework-specific.
These configs are committed to GitHub so the community can submit PRs
to improve them — ensuring fairness and transparency.

Usage:
  python3 generate-configs.py [--output-dir DIR]
"""

import os
import sys
from pathlib import Path

# =============================================================================
# Language-specific architecture patterns
# =============================================================================

# Base patterns (always included)
BASE_PATTERNS = [
    "*/types/*",
    "*/core/*",
    "*/internal/*",
]

# Language-specific architecture patterns
LANG_PATTERNS = {
    "go": [
        "*/router*.go",
        "*/middleware/*.go",
        "*/handler/*.go",
        "*/server/*.go",
        "*/config/*.go",
        "*/cmd/*.go",
        "**/interface*.go",
        "*_interface.go",
        "*/registry/*.go",
        "*/provider/*.go",
        "*/plugin/*.go",
        "*/api/*.go",
    ],
    "rust": [
        "*/lib.rs",
        "*/mod.rs",
        "*/traits/*.rs",
        "**/trait*.rs",
        "*/compiler/*.rs",
        "*/parser/*.rs",
        "*/ast/*.rs",
        "*/ir/*.rs",
        "*/codegen/*.rs",
        "*/typeck/*.rs",
        "*/resolve/*.rs",
        "*/hir/*.rs",
        "*/mir/*.rs",
    ],
    "c": [
        "*.h",
        "*/server.c",
        "*/networking.c",
        "*/aof.c",
        "*/rdb.c",
        "*/cluster.c",
        "*/replication.c",
        "*/module.c",
    ],
    "cpp": [
        "*.h",
        "*.hpp",
        "*/include/**/*.h",
        "*/source/**/factory*",
        "*/source/**/registry*",
        "*/source/**/config*",
        "*/api/*.h",
        "*/core/*.h",
    ],
    "javascript": [
        "*/router*.js",
        "*/middleware/*.js",
        "*/core/*.js",
        "*/lib/*.js",
        "*/plugin*.js",
        "*/config*.js",
        "*/index.js",
        "*/application.js",
    ],
    "typescript": [
        "*/router*.ts",
        "*/middleware/*.ts",
        "*/core/*.ts",
        "*/decorators/*.ts",
        "*/interfaces/*.ts",
        "**/interface*.ts",
        "*/plugin*.ts",
        "*/config*.ts",
        "*/factory*.ts",
        "*/module*.ts",
        "*/provider*.ts",
    ],
    "python": [
        "*/__init__.py",
        "*/routing*.py",
        "*/middleware*.py",
        "*/dependencies*.py",
        "*/config*.py",
        "*/core/*.py",
        "*/base*.py",
        "*/abc*.py",
        "*/protocols*.py",
    ],
    "java": [
        "**/interface*.java",
        "*Interface.java",
        "*/config/*.java",
        "*Configuration.java",
        "*/autoconfigure/*.java",
        "*/factory/*.java",
        "*Factory.java",
        "*/registry/*.java",
        "*/boot/*.java",
        "*/starter/*.java",
        "*/actuator/*.java",
    ],
    "elixir": [
        "*/router*.ex",
        "*/endpoint*.ex",
        "*/channel*.ex",
        "*/socket*.ex",
        "*/plug/*.ex",
        "*/controller*.ex",
        "*/live/*.ex",
        "*/config/*.exs",
        "mix.exs",
    ],
}

# Blame extensions per language
LANG_BLAME_EXT = {
    "go": ["*.go"],
    "rust": ["*.rs"],
    "c": ["*.c", "*.h"],
    "cpp": ["*.cpp", "*.hpp", "*.cc", "*.h"],
    "javascript": ["*.js", "*.mjs", "*.cjs"],
    "typescript": ["*.ts", "*.tsx"],
    "python": ["*.py"],
    "java": ["*.java"],
    "elixir": ["*.ex", "*.exs"],
}

# Per-repo exclude patterns
BASE_EXCLUDE = [
    "package-lock.json",
    "yarn.lock",
    "pnpm-lock.yaml",
    "go.sum",
    "*generated*",
    "*.gen.*",
    "*_test.go",
    "*_test.rs",
    "*_test.py",
    "test_*.py",
    "**/test/**",
    "**/tests/**",
    "**/testdata/**",
    "**/fixtures/**",
    "vendor/**",
    "node_modules/**",
    "*.pb.go",
    "*.pb.cc",
    "*.pb.h",
]

# =============================================================================
# Repository definitions
# =============================================================================

REPOS = {
    # Tier 1
    "react": {
        "lang": "javascript",
        "extra_blame_ext": ["*.ts", "*.tsx"],
        "extra_arch": [
            "packages/react/src/*.js",
            "packages/react-dom/src/*.js",
            "packages/react-reconciler/src/*.js",
            "packages/scheduler/src/*.js",
            "packages/shared/*.js",
        ],
        "extra_exclude": ["*.test.js", "*.test.ts", "**/__tests__/**"],
        "sample_size": 1000,
        "aliases": {
            "Dan Abramov": "Dan Abramov",
            "dan": "Dan Abramov",
            "gaearon": "Dan Abramov",
            "Sebastian Markbage": "Sebastian Markbage",
            "sebmarkbage": "Sebastian Markbage",
            "Andrew Clark": "Andrew Clark",
            "acdlite": "Andrew Clark",
            "Sophie Alpert": "Sophie Alpert",
            "sophiebits": "Sophie Alpert",
            "Ben Alpert": "Sophie Alpert",
            "Jordan Walke": "Jordan Walke",
            "jordwalke": "Jordan Walke",
        },
    },
    "kubernetes": {
        "lang": "go",
        "extra_arch": [
            "pkg/api/**/*.go",
            "staging/src/k8s.io/apimachinery/**/*.go",
            "staging/src/k8s.io/apiserver/**/*.go",
            "staging/src/k8s.io/client-go/**/*.go",
            "pkg/scheduler/*.go",
            "pkg/controller/**/*.go",
            "pkg/kubelet/*.go",
        ],
        "extra_exclude": ["*_test.go", "staging/**/fake/**", "hack/**"],
        "sample_size": 1500,
        "blame_timeout": 300,
        "aliases": {
            "Brendan Burns": "Brendan Burns",
            "brendandburns": "Brendan Burns",
            "Tim Hockin": "Tim Hockin",
            "thockin": "Tim Hockin",
            "Jordan Liggitt": "Jordan Liggitt",
            "liggitt": "Jordan Liggitt",
            "Clayton Coleman": "Clayton Coleman",
            "smarterclayton": "Clayton Coleman",
            "Daniel Smith": "Daniel Smith",
            "lavalamp": "Daniel Smith",
            "David Eads": "David Eads",
            "deads2k": "David Eads",
            "Kubernetes Submit Queue": "k8s-bot",
            "Kubernetes Prow Robot": "k8s-bot",
        },
        "extra_exclude_authors": ["k8s-bot", "Kubernetes Submit Queue", "Kubernetes Prow Robot", "k8s-merge-robot", "k8s-ci-robot"],
    },
    "terraform": {
        "lang": "go",
        "extra_arch": [
            "internal/command/*.go",
            "internal/terraform/*.go",
            "internal/configs/*.go",
            "internal/providers/*.go",
            "internal/backend/*.go",
            "internal/states/*.go",
        ],
        "extra_exclude": ["*_test.go"],
        "sample_size": 1000,
        "aliases": {
            "Mitchell Hashimoto": "Mitchell Hashimoto",
            "mitchellh": "Mitchell Hashimoto",
            "Martin Atkins": "Martin Atkins",
            "apparentlymart": "Martin Atkins",
            "James Bardin": "James Bardin",
            "jbardin": "James Bardin",
        },
    },
    "redis": {
        "lang": "c",
        "extra_arch": [
            "src/server.h",
            "src/server.c",
            "src/networking.c",
            "src/aof.c",
            "src/rdb.c",
            "src/cluster.c",
            "src/replication.c",
            "src/module.c",
            "src/dict.c",
            "src/sds.c",
            "src/ae.c",
        ],
        "sample_size": 500,
        "aliases": {
            "antirez": "Salvatore Sanfilippo",
            "Salvatore Sanfilippo": "Salvatore Sanfilippo",
            "Oran Agra": "Oran Agra",
            "oranagra": "Oran Agra",
        },
    },
    "rust": {
        "lang": "rust",
        "extra_arch": [
            "compiler/rustc/src/*.rs",
            "compiler/rustc_middle/src/**/*.rs",
            "compiler/rustc_typeck/src/**/*.rs",
            "compiler/rustc_resolve/src/**/*.rs",
            "compiler/rustc_codegen_llvm/src/**/*.rs",
            "library/core/src/**/*.rs",
            "library/std/src/**/*.rs",
        ],
        "extra_exclude": ["**/tests/**", "src/test/**"],
        "sample_size": 1500,
        "blame_timeout": 300,
        "aliases": {
            "Graydon Hoare": "Graydon Hoare",
            "graydon": "Graydon Hoare",
            "Niko Matsakis": "Niko Matsakis",
            "nikomatsakis": "Niko Matsakis",
            "Aaron Turon": "Aaron Turon",
            "aturon": "Aaron Turon",
            "Felix S Klock II": "Felix Klock",
            "Felix Klock": "Felix Klock",
            "pnkfelix": "Felix Klock",
            "Eduard-Mihai Burtescu": "Eduard-Mihai Burtescu",
            "eddyb": "Eduard-Mihai Burtescu",
            "Alex Crichton": "Alex Crichton",
            "alexcrichton": "Alex Crichton",
            "Manish Goregaokar": "Manish Goregaokar",
            "Manishearth": "Manish Goregaokar",
        },
        "extra_exclude_authors": ["bors", "bors[bot]", "rust-timer", "rust-log-analyzer", "rustbot"],
    },
    # Tier 2: Infrastructure
    "prometheus": {
        "lang": "go",
        "extra_arch": [
            "tsdb/**/*.go",
            "promql/**/*.go",
            "scrape/*.go",
            "storage/**/*.go",
            "rules/*.go",
            "web/**/*.go",
        ],
        "sample_size": 800,
    },
    "grafana": {
        "lang": "typescript",
        "extra_blame_ext": ["*.go"],
        "extra_arch": [
            "pkg/api/*.go",
            "pkg/services/**/*.go",
            "packages/grafana-data/src/**/*.ts",
            "packages/grafana-ui/src/**/*.ts",
            "public/app/core/**/*.ts",
        ],
        "sample_size": 1000,
    },
    "loki": {
        "lang": "go",
        "extra_arch": [
            "pkg/storage/**/*.go",
            "pkg/querier/**/*.go",
            "pkg/ingester/**/*.go",
            "pkg/distributor/**/*.go",
        ],
        "sample_size": 800,
    },
    "argo-cd": {
        "lang": "go",
        "extra_arch": [
            "controller/**/*.go",
            "server/**/*.go",
            "reposerver/**/*.go",
            "util/**/*.go",
        ],
        "sample_size": 800,
    },
    "envoy": {
        "lang": "cpp",
        "extra_arch": [
            "source/common/**/*.h",
            "source/server/**/*.h",
            "source/extensions/**/*.h",
            "include/envoy/**/*.h",
        ],
        "extra_exclude": ["**/*_test.cc"],
        "sample_size": 1000,
        "blame_timeout": 300,
    },
    # Tier 2: Backend Frameworks
    "fastapi": {
        "lang": "python",
        "extra_arch": [
            "fastapi/routing.py",
            "fastapi/applications.py",
            "fastapi/dependencies/**/*.py",
            "fastapi/security/**/*.py",
        ],
        "sample_size": 500,
        "aliases": {
            "tiangolo": "Sebastian Ramirez",
            "Sebastian Ramirez": "Sebastian Ramirez",
            "Sebastián Ramírez": "Sebastian Ramirez",
        },
    },
    "nest": {
        "lang": "typescript",
        "extra_arch": [
            "packages/core/src/**/*.ts",
            "packages/common/src/decorators/**/*.ts",
            "packages/common/src/interfaces/**/*.ts",
            "packages/microservices/src/**/*.ts",
        ],
        "extra_exclude": ["*.spec.ts"],
        "sample_size": 800,
        "aliases": {
            "Kamil Mysliwiec": "Kamil Mysliwiec",
            "kamilmysliwiec": "Kamil Mysliwiec",
        },
    },
    "spring-boot": {
        "lang": "java",
        "extra_arch": [
            "spring-boot-project/spring-boot/src/main/**/*.java",
            "spring-boot-project/spring-boot-autoconfigure/src/main/**/*.java",
            "spring-boot-project/spring-boot-actuator/src/main/**/*.java",
        ],
        "extra_exclude": ["*Test.java", "*Tests.java"],
        "sample_size": 1000,
        "aliases": {
            "Phillip Webb": "Phillip Webb",
            "philwebb": "Phillip Webb",
            "Andy Wilkinson": "Andy Wilkinson",
            "wilkinsona": "Andy Wilkinson",
            "Stephane Nicoll": "Stephane Nicoll",
            "snicoll": "Stephane Nicoll",
        },
    },
    "express": {
        "lang": "javascript",
        "extra_arch": [
            "lib/application.js",
            "lib/router/*.js",
            "lib/middleware/*.js",
            "lib/request.js",
            "lib/response.js",
        ],
        "sample_size": 500,
        "aliases": {
            "Tj Holowaychuk": "TJ Holowaychuk",
            "TJ Holowaychuk": "TJ Holowaychuk",
            "tj": "TJ Holowaychuk",
            "visionmedia": "TJ Holowaychuk",
            "Douglas Christopher Wilson": "Douglas Christopher Wilson",
            "dougwilson": "Douglas Christopher Wilson",
        },
    },
    "phoenix": {
        "lang": "elixir",
        "extra_arch": [
            "lib/phoenix/router*.ex",
            "lib/phoenix/endpoint*.ex",
            "lib/phoenix/channel*.ex",
            "lib/phoenix/socket*.ex",
            "lib/phoenix/controller*.ex",
        ],
        "sample_size": 500,
        "aliases": {
            "Chris McCord": "Chris McCord",
            "chrismccord": "Chris McCord",
            "José Valim": "Jose Valim",
            "Jose Valim": "Jose Valim",
            "josevalim": "Jose Valim",
        },
    },
    # Tier 2: Developer Tools
    "esbuild": {
        "lang": "go",
        "extra_arch": [
            "pkg/api/*.go",
            "pkg/js_parser/*.go",
            "pkg/css_parser/*.go",
            "pkg/js_ast/*.go",
            "pkg/bundler/*.go",
        ],
        "sample_size": 500,
        "aliases": {
            "Evan Wallace": "Evan Wallace",
            "evanw": "Evan Wallace",
        },
    },
    "swc": {
        "lang": "rust",
        "extra_arch": [
            "crates/swc_ecma_parser/src/**/*.rs",
            "crates/swc_ecma_transforms*/src/**/*.rs",
            "crates/swc_ecma_ast/src/**/*.rs",
            "crates/swc_core/src/**/*.rs",
        ],
        "sample_size": 800,
        "aliases": {
            "강동윤": "Donny (kdy1)",
            "Donny/강동윤": "Donny (kdy1)",
            "kdy1": "Donny (kdy1)",
            "Donny": "Donny (kdy1)",
        },
    },
    "vite": {
        "lang": "typescript",
        "extra_arch": [
            "packages/vite/src/node/**/*.ts",
            "packages/vite/src/node/server/**/*.ts",
            "packages/vite/src/node/plugins/**/*.ts",
            "packages/vite/src/node/optimizer/**/*.ts",
        ],
        "extra_exclude": ["*.spec.ts"],
        "sample_size": 500,
        "aliases": {
            "Evan You": "Evan You",
            "yyx990803": "Evan You",
        },
    },
    "prettier": {
        "lang": "javascript",
        "extra_arch": [
            "src/language-*/index.js",
            "src/language-*/parser*.js",
            "src/language-*/printer*.js",
            "src/main/*.js",
            "src/common/*.js",
        ],
        "sample_size": 500,
    },
    "eslint": {
        "lang": "javascript",
        "extra_arch": [
            "lib/linter/*.js",
            "lib/rules/*.js",
            "lib/config/*.js",
            "lib/rule-tester/*.js",
            "lib/source-code/**/*.js",
        ],
        "sample_size": 500,
    },
    # Tier 2: Data / Systems
    "duckdb": {
        "lang": "cpp",
        "extra_arch": [
            "src/include/**/*.hpp",
            "src/execution/**/*.cpp",
            "src/optimizer/**/*.cpp",
            "src/planner/**/*.cpp",
            "src/storage/**/*.cpp",
            "src/parser/**/*.cpp",
            "src/catalog/**/*.cpp",
        ],
        "sample_size": 1000,
    },
    "ClickHouse": {
        "lang": "cpp",
        "extra_arch": [
            "src/Interpreters/**/*.h",
            "src/Storages/**/*.h",
            "src/Parsers/**/*.h",
            "src/Processors/**/*.h",
            "src/Server/**/*.h",
            "src/Core/**/*.h",
        ],
        "extra_exclude": ["**/*Test*", "**/gtest*"],
        "sample_size": 1500,
        "blame_timeout": 300,
    },
    "arrow": {
        "lang": "cpp",
        "extra_blame_ext": ["*.py", "*.rs", "*.java"],
        "extra_arch": [
            "cpp/src/arrow/*.h",
            "cpp/src/arrow/compute/**/*.h",
            "cpp/src/arrow/io/**/*.h",
            "cpp/src/parquet/**/*.h",
            "python/pyarrow/*.py",
        ],
        "sample_size": 1000,
        "blame_timeout": 300,
    },
    "polars": {
        "lang": "rust",
        "extra_arch": [
            "crates/polars-core/src/**/*.rs",
            "crates/polars-lazy/src/**/*.rs",
            "crates/polars-plan/src/**/*.rs",
            "crates/polars-expr/src/**/*.rs",
            "crates/polars-io/src/**/*.rs",
        ],
        "sample_size": 800,
        "aliases": {
            "Ritchie Vink": "Ritchie Vink",
            "ritchie46": "Ritchie Vink",
        },
    },
    "superset": {
        "lang": "python",
        "extra_blame_ext": ["*.ts", "*.tsx"],
        "extra_arch": [
            "superset/views/**/*.py",
            "superset/models/**/*.py",
            "superset/security/**/*.py",
            "superset/config.py",
            "superset-frontend/src/dashboard/**/*.ts",
        ],
        "sample_size": 1000,
    },
}


def generate_config(name: str, spec: dict) -> str:
    """Generate eis.yaml content for a repository."""
    lang = spec["lang"]
    sample_size = spec.get("sample_size", 500)
    blame_timeout = spec.get("blame_timeout", 120)

    # Merge architecture patterns
    arch_patterns = BASE_PATTERNS.copy()
    arch_patterns.extend(LANG_PATTERNS.get(lang, []))
    arch_patterns.extend(spec.get("extra_arch", []))

    # Merge blame extensions
    blame_ext = LANG_BLAME_EXT.get(lang, ["*.go"]).copy()
    blame_ext.extend(spec.get("extra_blame_ext", []))
    # Deduplicate
    blame_ext = list(dict.fromkeys(blame_ext))

    # Merge exclude patterns
    exclude = BASE_EXCLUDE.copy()
    exclude.extend(spec.get("extra_exclude", []))

    # Generate YAML
    lines = [
        f"# EIS config for {name}",
        f"# Language: {lang}",
        f"#",
        f"# Architecture patterns define which files represent structural decisions.",
        f"# Submit a PR to improve these patterns for fairer analysis.",
        f"# See: https://github.com/machuz/engineering-impact-score/tree/main/research/oss-gravity-map",
        f"",
        f"sample_size: {sample_size}",
        f"blame_timeout: {blame_timeout}",
        f"",
        f"exclude_file_patterns:",
    ]
    for p in exclude:
        lines.append(f'  - "{p}"')

    lines.append("")
    lines.append("architecture_patterns:")
    for p in arch_patterns:
        lines.append(f'  - "{p}"')

    lines.append("")
    lines.append("blame_extensions:")
    for ext in blame_ext:
        lines.append(f'  - "{ext}"')

    # Aliases (git author name → canonical name)
    aliases = spec.get("aliases", {})
    if aliases:
        lines.append("")
        lines.append("aliases:")
        for from_name, to_name in aliases.items():
            if from_name != to_name:  # Only write actual aliases, not identity mappings
                lines.append(f'  "{from_name}": "{to_name}"')

    # Exclude authors
    exclude_authors = [
        "github-actions[bot]",
        "renovate[bot]",
        "dependabot[bot]",
        "web-flow",
        "bors[bot]",
        "rust-timer",
        "rust-log-analyzer",
    ]
    exclude_authors.extend(spec.get("extra_exclude_authors", []))
    # Deduplicate
    exclude_authors = list(dict.fromkeys(exclude_authors))

    lines.append("")
    lines.append("exclude_authors:")
    for author in exclude_authors:
        lines.append(f'  - "{author}"')
    lines.append("")

    return "\n".join(lines) + "\n"


def main():
    output_dir = sys.argv[1] if len(sys.argv) > 1 else "configs"
    os.makedirs(output_dir, exist_ok=True)

    for name, spec in REPOS.items():
        config = generate_config(name, spec)
        path = os.path.join(output_dir, f"{name}.yaml")
        with open(path, "w") as f:
            f.write(config)
        print(f"  {name}.yaml ({spec['lang']})")

    print(f"\nGenerated {len(REPOS)} configs in {output_dir}/")


if __name__ == "__main__":
    main()
