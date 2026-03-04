// main.go is the AGD CLI entrypoint and top-level command router.
// main.go는 AGD CLI 진입점과 상위 명령 라우팅을 담당합니다.
package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		return commandWizard(nil)
	}

	switch args[0] {
	case "help", "--help", "-h":
		return commandHelp(args[1:])
	case "quick":
		return commandHelp(args[1:])
	case "wizard", "start":
		return commandWizard(args[1:])
	case "new":
		return commandNewEasy(args[1:])
	case "check":
		return commandCheckEasy(args[1:])
	case "check-all", "check-tree", "audit":
		return commandCheckAllEasy(args[1:])
	case "search":
		return commandSearchEasy(args[1:])
	case "edit":
		return commandEditEasy(args[1:])
	case "add":
		return commandAddEasy(args[1:])
	case "section-add", "add-section":
		return commandSectionAddEasy(args[1:])
	case "map-sync", "sync-map":
		return commandMapSyncEasy(args[1:])
	case "map-suggest", "source-map":
		return commandMapSuggestEasy(args[1:])
	case "logic-log", "change-log":
		return commandLogicLogEasy(args[1:])
	case "incident-tag", "feature-tag", "tag-feature":
		return commandIncidentTagEasy(args[1:])
	case "role-set", "authority-set":
		return commandRoleSetEasy(args[1:])
	case "role-graph", "graph", "relation-graph":
		return commandRoleGraphEasy(args[1:])
	case "kit",
		"starter-kit", "maintenance", "maintenance-kit",
		"bridge-lite", "bridge", "bridge-kit",
		"new-project", "new-project-kit",
		"change-flow",
		"incident-response", "incident", "incident-kit",
		"incident-lifecycle",
		"quality-gate":
		return commandKitEasy(append([]string{args[0]}, args[1:]...))
	case "view":
		return commandViewEasy(args[1:])
	case "lint":
		return commandLint(args[1:])
	case "find":
		return commandFind(args[1:])
	case "patch":
		return commandPatch(args[1:])
	case "export":
		return commandExport(args[1:])
	case "init":
		return commandInit(args[1:])
	default:
		fmt.Fprintf(os.Stderr, text("unknown command %q\n\n", "알 수 없는 명령어 %q\n\n"), args[0])
		fmt.Fprint(os.Stderr, usage())
		return 2
	}
}

func usage() string {
	return text(`AGD CLI

Usage:
  agd                     (start interactive wizard)
  agd quick
  agd wizard
  agd new <type> <out.agd> [title] [owner]
  agd check <file.agd>
  agd check-all [root] [--strict] [--include-archive]
  agd search <file.agd> <keyword>
  agd edit <file.agd> <SECTION-ID> <new-summary> --reason "<why>" --impact "<effect>"
  agd add <file.agd> <SECTION-ID> <text-to-append> --reason "<why>" --impact "<effect>"
  agd section-add <file.agd> <SECTION-ID> <title> <summary> [links] [content] --reason "<why>" --impact "<effect>"
  agd map-sync <file.agd> [--reason "<why>" --impact "<effect>"]
  agd map-suggest <derived-file.agd> <source-doc.agd> [auto|strict-auto|smart-auto]
  agd logic-log <file.agd> <SECTION-ID> --reason "<why>" --impact "<effect>"
  agd incident-tag <target.agd> <feature-tag> [source-doc.agd] [section-id] [--module "<name>"] --reason "<why>" --impact "<effect>"
  agd role-set <file.agd> <source|derived> [source-doc] [source-sections]
  agd role-graph [root] [--format text|mermaid] [--out file] [--include-archive]
  agd kit <starter-kit|new-project|maintenance|incident> <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]
  agd starter-kit|new-project|maintenance|incident <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]
  agd view <file.agd> [out.md]

Advanced:
  agd lint <file.agd>
  agd find <file.agd> <query>
  agd patch <file.agd> --section <SEC-ID> [--title "..."] [--summary "..."] [--set-content "..."] [--append-content "..."] [--links "A,B,C"] --reason "..." --impact "..." [--author "..."] [--out file.agd]
  agd export <file.agd> --format md [--out file.md]
  agd init --template <name> --out <file.agd> [--title "..."] [--owner "..."] [--version "..."] [--force]
  agd init --list
`, `AGD CLI

사용법:
  agd                     (대화형 위자드 시작)
  agd quick
  agd wizard
  agd new <type> <out.agd> [title] [owner]
  agd check <file.agd>
  agd check-all [root] [--strict] [--include-archive]
  agd search <file.agd> <keyword>
  agd edit <file.agd> <SECTION-ID> <new-summary> --reason "<why>" --impact "<effect>"
  agd add <file.agd> <SECTION-ID> <text-to-append> --reason "<why>" --impact "<effect>"
  agd section-add <file.agd> <SECTION-ID> <title> <summary> [links] [content] --reason "<why>" --impact "<effect>"
  agd map-sync <file.agd> [--reason "<why>" --impact "<effect>"]
  agd map-suggest <derived-file.agd> <source-doc.agd> [auto|strict-auto|smart-auto]
  agd logic-log <file.agd> <SECTION-ID> --reason "<why>" --impact "<effect>"
  agd incident-tag <target.agd> <feature-tag> [source-doc.agd] [section-id] [--module "<name>"] --reason "<why>" --impact "<effect>"
  agd role-set <file.agd> <source|derived> [source-doc] [source-sections]
  agd role-graph [root] [--format text|mermaid] [--out file] [--include-archive]
  agd kit <starter-kit|new-project|maintenance|incident> <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]
  agd starter-kit|new-project|maintenance|incident <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]
  agd view <file.agd> [out.md]

고급:
  agd lint <file.agd>
  agd find <file.agd> <query>
  agd patch <file.agd> --section <SEC-ID> [--title "..."] [--summary "..."] [--set-content "..."] [--append-content "..."] [--links "A,B,C"] --reason "..." --impact "..." [--author "..."] [--out file.agd]
  agd export <file.agd> --format md [--out file.md]
  agd init --template <name> --out <file.agd> [--title "..."] [--owner "..."] [--version "..."] [--force]
  agd init --list
`)
}
