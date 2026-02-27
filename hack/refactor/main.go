// Expense Tracker Bot — Automated Refactor Script
//
// Run from the repository root of the 'natural' branch:
//
//	go run refactor.go
//
// Flags:
//
//	--dry-run   Preview every change without touching any file
//	--verbose   Print each individual substitution as it happens
//	--skip      Comma-separated phases to skip:
//	            renames, types, fields, newfiles, makefile, docker, ci, config
//
// Example (preview only):
//
//	go run refactor.go --dry-run --verbose
//
// Example (skip Makefile + Docker):
//
//	go run refactor.go --skip=makefile,docker

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ─── Terminal colours ────────────────────────────────────────────────────────

const (
	cReset  = "\033[0m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cCyan   = "\033[36m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
)

func red(s string) string { return cRed + s + cReset }

func green(s string) string { return cGreen + s + cReset }

func yellow(s string) string { return cYellow + s + cReset }

func blue(s string) string { return cBlue + s + cReset }

func cyan(s string) string { return cCyan + s + cReset }

func bold(s string) string { return cBold + s + cReset }

func dim(s string) string { return cDim + s + cReset }

// ─── Globals ──────────────────────────────────────────────────────────────────

var (
	dryRun  *bool
	verbose *bool
	skipArg *string

	rootDir string

	// Tracking
	renamedFiles  []string
	modifiedFiles []string
	createdFiles  []string
	skippedFiles  []string
	warnings      []string
	manualSteps   []string
)

// ─── Entry point ──────────────────────────────────────────────────────────────

func main() {
	dryRun = flag.Bool("dry-run", false, "Preview changes without modifying any file")
	verbose = flag.Bool("verbose", false, "Print each substitution as it happens")
	skipArg = flag.String("skip", "", "Comma-separated phases to skip")
	flag.Parse()

	rootDir = "."
	if args := flag.Args(); len(args) > 0 {
		rootDir = args[0]
	}

	printBanner()

	if err := verifyProject(); err != nil {
		fatalf("Could not verify project root: %v\n"+
			"Make sure you run this from the expense-tracker-bot repo root.\n", err)
	}

	if !*dryRun {
		must(createBackup(), "create backup")
	} else {
		infof("%s  No files will be changed.\n\n", yellow("[DRY-RUN]"))
	}

	skip := parseSkip(*skipArg)

	runPhase("1 · File renames", !skip["renames"], phaseFileRenames)
	runPhase("2 · Type substitutions", !skip["types"], phaseTypeSubstitutions)
	runPhase("3 · Struct field additions", !skip["fields"], phaseStructFields)
	runPhase("4 · New source files", !skip["newfiles"], phaseNewFiles)
	runPhase("5 · Makefile", !skip["makefile"], phaseMakefile)
	runPhase("6 · Dockerfile", !skip["docker"], phaseDockerfile)
	runPhase("7 · CI/CD workflows", !skip["ci"], phaseCICD)
	runPhase("8 · Config & lint files", !skip["config"], phaseConfigFiles)

	printSummary()
}

// ════════════════════════════════════════════════════════════════════════════
// PHASE 1 — File renames
// ════════════════════════════════════════════════════════════════════════════

func phaseFileRenames() {
	// ── 1. Flat file renames (single files, no directory change) ─────────────
	flat := []struct{ from, to string }{
		// Top-level repo interface files
		{"repos/accounts.go", "repos/wallets.go"},
		// Top-level service interface files
		{"services/accounts.go", "services/wallets.go"},
		// API handler
		{"api/handlers/wallet.go", "api/handlers/wallet.go"},
		// contact (debtor/creditor) implementation files → contact
		{"repos/user/contact.go", "repos/user/contact.go"},
		{"services/user/contact.go", "services/user/contact.go"},
	}
	for _, r := range flat {
		renameFile(r.from, r.to)
	}

	// ── 2. Directory renames (rename whole sub-directory + file inside) ───────
	// repos/accounts/ → repos/wallets/  (file inside: accounts.go → wallets.go)
	renameDir("repos/accounts", "repos/wallets", map[string]string{
		"accounts.go": "wallets.go",
	})
	// services/accounts/ → services/wallets/  (file inside: accounts.go → wallets.go)
	renameDir("services/accounts", "services/wallets", map[string]string{
		"accounts.go": "wallets.go",
	})

	// ── 3. models/user.go — smart rename based on content ────────────────────
	// The file holds the bot-user (Profile) struct. The debtor/creditor types
	// live in repos/user/contact.go and services/user/contact.go, not in models.
	// Rename models/user.go → models/profile.go unconditionally; emit a
	// warning if the file also contains Balance/debtor fields so the user
	// knows to manually extract a Contact model.
	userModelPath := filepath.Join(rootDir, "models", "user.go")
	if src, err := os.ReadFile(userModelPath); err == nil {
		content := string(src)
		hasDebtor := strings.Contains(content, "Contact") ||
			strings.Contains(content, "debtor") ||
			strings.Contains(content, "Contact") ||
			strings.Contains(content, "contact") ||
			strings.Contains(content, "Contact")
		renameFile("models/user.go", "models/profile.go")
		if hasDebtor {
			warnings = append(warnings,
				"models/user.go (now profile.go) appears to contain debtor/creditor fields. "+
					"Review models/profile.go and extract Contact-related fields into a new models/contact.go if needed.")
		}
	}

	// ── 4. gqtypes — type subs only, no file rename needed ───────────────────
	// models/gqtypes/user.go stays named user.go; type substitutions in Phase 2
	// will update the struct/type names inside it.
}

// renameFile renames a single file relative to rootDir.
// Silently skips if the source does not exist.
func renameFile(from, to string) {
	src := filepath.Join(rootDir, from)
	dst := filepath.Join(rootDir, to)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		if *verbose {
			infof("  %s  %s (not found, skip)\n", dim("SKIP"), from)
		}
		return
	}
	if src == dst {
		return
	}
	logAction("RENAME", from+" → "+to)
	if !*dryRun {
		must(os.MkdirAll(filepath.Dir(dst), 0o755), "mkdir for "+to)
		must(os.Rename(src, dst), "rename "+from)
		renamedFiles = append(renamedFiles, from+" → "+to)
	}
}

// renameDir renames a sub-directory and optionally renames individual files
// within it. fileMap maps old basename → new basename; files not in the map
// keep their original name.
func renameDir(fromRel, toRel string, fileMap map[string]string) {
	srcDir := filepath.Join(rootDir, fromRel)
	dstDir := filepath.Join(rootDir, toRel)

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		if *verbose {
			infof("  %s  %s/ (dir not found, skip)\n", dim("SKIP"), fromRel)
		}
		return
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Could not read directory %s: %v", fromRel, err))
		return
	}

	logAction("RENAME DIR", fromRel+"/ → "+toRel+"/")
	if *dryRun {
		for _, e := range entries {
			newName := e.Name()
			if mapped, ok := fileMap[e.Name()]; ok {
				newName = mapped
			}
			infof("  %s  %s/%s → %s/%s\n", dim("  would move"), fromRel, e.Name(), toRel, newName)
		}
		return
	}

	must(os.MkdirAll(dstDir, 0o755), "mkdir "+toRel)
	for _, e := range entries {
		if e.IsDir() {
			continue // only move regular files at this level
		}
		newName := e.Name()
		if mapped, ok := fileMap[e.Name()]; ok {
			newName = mapped
		}
		src := filepath.Join(srcDir, e.Name())
		dst := filepath.Join(dstDir, newName)
		must(os.Rename(src, dst), "move "+e.Name())
		renamedFiles = append(renamedFiles, fromRel+"/"+e.Name()+" → "+toRel+"/"+newName)
	}
	// Remove now-empty source dir (best-effort)
	_ = os.Remove(srcDir)
}

// ════════════════════════════════════════════════════════════════════════════
// PHASE 2 — Type substitutions across all .go files
// ════════════════════════════════════════════════════════════════════════════

// sub describes one regex find-and-replace with an explanation.
type sub struct {
	pattern string // Go regexp (applied with word-boundary anchors automatically when wb=true)
	replace string
	wb      bool // wrap pattern in \b word-boundaries
	desc    string
}

func phaseTypeSubstitutions() {
	// ORDER MATTERS — longer / more specific patterns must come first to
	// avoid partial replacements (e.g. WalletRepo before Wallet).
	subs := []sub{
		// ── Wallet → Wallet (types, interfaces, constructors) ───────────────
		{"WalletType", "WalletType", true, "WalletType → WalletType"},
		{"WalletRepo", "WalletRepo", true, "WalletRepo → WalletRepo"},
		{"WalletService", "WalletService", true, "WalletService → WalletService"},
		{"NewWalletRepo", "NewWalletRepo", true, "NewWalletRepo → NewWalletRepo"},
		{"NewWalletService", "NewWalletService", true, "NewWalletService → NewWalletService"},
		{"NewWallet", "NewWallet", true, "NewWallet constructor → NewWallet"},
		{"newWallet", "newWallet", true, "newWallet → newWallet"},
		{"Wallet", "Wallet", true, "Wallet struct/type → Wallet"},
		// lower-case variable names (trailing context avoids spurious matches)
		{"wallet ", "wallet ", false, "var 'wallet ' → 'wallet '"},
		{"wallet,", "wallet,", false, "var 'wallet,' → 'wallet,'"},
		{"wallet)", "wallet)", false, "var 'wallet)' → 'wallet)'"},
		{"wallet.", "wallet.", false, "var 'wallet.' → 'wallet.'"},
		{"wallet{", "wallet{", false, "var 'wallet{' → 'wallet{'"},
		{":= wallet", ":= wallet", false, ":= wallet → := wallet"},

		// ── Import paths and package declarations for accounts → wallets ─────
		// These handle: import ".../repos/wallets" and: package wallets
		{"/repos/accounts\"", "/repos/wallets\"", false, "import path repos/accounts → repos/wallets"},
		{"/services/accounts\"", "/services/wallets\"", false, "import path services/accounts → services/wallets"},
		{"package wallets", "package wallets", false, "package decl accounts → wallets"},

		// ── contact (debtor/creditor) → contact ─────────────────────────────────
		// "contact" is the project-specific shorthand for debtor/creditor
		{"ContactRepo", "ContactRepo", true, "ContactRepo → ContactRepo"},
		{"ContactService", "ContactService", true, "ContactService → ContactService"},
		{"NewContact", "NewContact", true, "NewContact → NewContact"},
		{"Contact", "Contact", true, "Contact struct/type → Contact"},
		{"contact", "contact", true, "contact var/pkg → contact"},
		// Also cover any spelled-out forms that may exist
		{"Contacts", "Contacts", true, "Contacts → Contacts"},
		{"Contact", "Contact", true, "Contact → Contact"},
		{"ContactRepo", "ContactRepo", true, "ContactRepo → ContactRepo"},
		{"ContactService", "ContactService", true, "ContactService → ContactService"},
		{"NewContact", "NewContact", true, "NewContact → NewContact"},
		{"Contact", "Contact", true, "Contact struct/type → Contact"},
		{"Contact", "Contact", true, "Contact → Contact"},

		// ── User (bot owner) → Profile — ONLY where Profile prefix is used ───
		// Plain "User" is handled per-file below to avoid clobbering
		// third-party type names (telebot.User, etc.)
		{"ProfileRepo", "ProfileRepo", true, "ProfileRepo → ProfileRepo"},
		{"ProfileService", "ProfileService", true, "ProfileService → ProfileService"},
		{"Profile", "Profile", true, "Profile → Profile"},

		// ── Telegram bot command strings ──────────────────────────────────────
		{`"/contacts"`, `"/contacts"`, false, `bot command /users → /contacts`},
		{`"Contacts"`, `"Contacts"`, false, `menu label Contacts → Contacts`},
		{`"contact"`, `"contact"`, false, `label "contact" → "contact"`},

		// ── Menu / keyboard label strings ─────────────────────────────────────
		// Use exact quoted forms to avoid touching unrelated identifiers
		{`"Wallet"`, `"Wallet"`, false, `menu label "Wallet" → "Wallet"`},
		{`"Wallets"`, `"Wallets"`, false, `menu label "Wallets" → "Wallets"`},
	}

	// Walk every .go file in the project (excluding vendor/)
	must(walkGoFiles(func(path string) error {
		return applySubstitutions(path, subs)
	}), "walk go files for substitutions")

	// ── Profile-specific pass ─────────────────────────────────────────────────
	// Rename plain "User" → "Profile" only in files that are definitely about
	// the bot owner, to avoid breaking telebot.User and similar external types.
	profileFiles := []string{
		// model file (renamed from models/user.go in Phase 1)
		filepath.Join(rootDir, "models", "profile.go"),
		// gqtypes may have a User GraphQL type for the bot owner
		filepath.Join(rootDir, "models", "gqtypes", "user.go"),
		// repo interface + implementation for the bot-owner user
		filepath.Join(rootDir, "repos", "user.go"),
		filepath.Join(rootDir, "repos", "user", "user.go"),
		// service interface + implementation
		filepath.Join(rootDir, "services", "user.go"),
		filepath.Join(rootDir, "services", "user", "user.go"),
		// aggregator service — imports and wires all sub-services including user
		filepath.Join(rootDir, "services", "all", "all.go"),
		// API handler for user-facing commands (/start, /profile, etc.)
		filepath.Join(rootDir, "api", "handlers", "user.go"),
	}
	profileSubs := []sub{
		{"type User struct", "type Profile struct", false, "type User struct → type Profile struct"},
		{"models.User{", "models.Profile{", false, "models.User{ → models.Profile{"},
		{"models.User)", "models.Profile)", false, "models.User) → models.Profile)"},
		{"[]models.User", "[]models.Profile", false, "[]models.User → []models.Profile"},
		{"*models.User", "*models.Profile", false, "*models.User → *models.Profile"},
		{"models.User ", "models.Profile ", false, "models.User  → models.Profile "},
		// Interface / service type names
		{"UserRepo", "ProfileRepo", true, "UserRepo → ProfileRepo (profile files)"},
		{"UserService", "ProfileService", true, "UserService → ProfileService (profile files)"},
		{"NewUserRepo", "NewProfileRepo", true, "NewUserRepo → NewProfileRepo"},
		{"NewUserService", "NewProfileService", true, "NewUserService → NewProfileService"},
	}
	for _, pf := range profileFiles {
		if _, err := os.Stat(pf); err == nil {
			must(applySubstitutions(pf, profileSubs), "profile subs in "+pf)
		}
	}

	// ── Contact-specific pass ─────────────────────────────────────────────────
	// Apply contact-only subs to the contact implementation files that were
	// renamed to contact.go in Phase 1, plus any model file that has those types.
	contactFiles := []string{
		filepath.Join(rootDir, "repos", "user", "contact.go"),
		filepath.Join(rootDir, "services", "user", "contact.go"),
		filepath.Join(rootDir, "models", "gqtypes", "user.go"),
		// If contact types ended up in profile.go, cover that too
		filepath.Join(rootDir, "models", "profile.go"),
	}
	contactSubs := []sub{
		// If the contact file still has a User struct for the contact, rename it
		{"type User struct", "type Contact struct", false, "type User struct → type Contact struct"},
		{"models.User{", "models.Contact{", false, "models.User{ → models.Contact{"},
		{"[]models.User", "[]models.Contact", false, "[]models.User → []models.Contact"},
		{"*models.User", "*models.Contact", false, "*models.User → *models.Contact"},
		// Rename the Balance field to NetBalance on the Contact model
		// (scoped to these files only to avoid hitting Transaction.Balance etc.)
		{"\tBalance ", "\tNetBalance ", false, "Balance field → NetBalance (contact)"},
	}
	for _, cf := range contactFiles {
		if _, err := os.Stat(cf); err == nil {
			must(applySubstitutions(cf, contactSubs), "contact subs in "+cf)
		}
	}

	// ── Wallet package-declaration pass ──────────────────────────────────────
	// After renameDir the file is at repos/wallets/wallets.go with `package wallets`.
	// The global sub above handles `package wallets → package wallets` but only
	// if the sub ran before this point — which it did. This is a belt-and-suspenders
	// explicit pass in case the file had a non-standard package name.
	walletImplFiles := []string{
		filepath.Join(rootDir, "repos", "wallets", "wallets.go"),
		filepath.Join(rootDir, "services", "wallets", "wallets.go"),
	}
	walletDeclSubs := []sub{
		{"package wallets", "package wallets", false, "package decl → wallets"},
		{"package wallets\n", "package wallets\n", false, "no-op (already correct)"},
	}
	for _, wf := range walletImplFiles {
		if _, err := os.Stat(wf); err == nil {
			must(applySubstitutions(wf, walletDeclSubs), "wallet pkg decl in "+wf)
		}
	}
}

// applySubstitutions applies a list of subs to a single file.
func applySubstitutions(path string, subs []sub) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	original := string(raw)
	content := original

	for _, s := range subs {
		if s.pattern == s.replace {
			continue // skip no-ops
		}
		var re *regexp.Regexp
		if s.wb {
			re = regexp.MustCompile(`\b` + regexp.QuoteMeta(s.pattern) + `\b`)
		} else {
			re = regexp.MustCompile(regexp.QuoteMeta(s.pattern))
		}
		updated := re.ReplaceAllString(content, s.replace)
		if updated != content {
			if *verbose {
				infof("  %s  %-60s  %s\n", cyan("SUB"), dim(filepath.Base(path)), s.desc)
			}
			content = updated
		}
	}

	if content == original {
		return nil
	}

	logAction("MODIFY", path)
	if !*dryRun {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
		modifiedFiles = appendUniq(modifiedFiles, path)
	}
	return nil
}

// ════════════════════════════════════════════════════════════════════════════
// PHASE 3 — Struct field additions
// ════════════════════════════════════════════════════════════════════════════

type fieldAdd struct {
	file         string // relative to root, supports globs
	structName   string
	field        string // full field line, e.g. `\tDeletedAt *time.Time \`db:"deleted_at"\``
	importNeeded string // import path to add if field uses an external type
}

func phaseStructFields() {
	// ── Locate where the Wallet/Wallet struct actually lives ─────────────────
	// The project has no models/wallet.go; the Wallet struct may be in
	// models/balance.go, models/expense.go, or another model file.
	// We scan all model files and add the Version field to whichever one
	// contains `type Wallet struct` (or `type Wallet struct` after Phase 2).
	walletModelFile := findStructInModels("Wallet", "Wallet")
	if walletModelFile == "" {
		warnings = append(warnings,
			"Could not find 'type Wallet struct' or 'type Wallet struct' in models/. "+
				"Add 'Version int' (optimistic lock) to that struct manually.")
	}

	additions := []fieldAdd{
		// ── Transaction: soft-delete support ─────────────────────────────────
		{
			file:         "models/transaction.go",
			structName:   "Transaction",
			field:        "\tDeletedAt  *time.Time `db:\"deleted_at\"` // nil = active; non-nil = soft-deleted",
			importNeeded: "time",
		},
		{
			file:         "models/transaction.go",
			structName:   "Transaction",
			field:        "\tCreatedAt  time.Time  `db:\"created_at\"`",
			importNeeded: "time",
		},
		// ── Profile (bot owner): timezone ────────────────────────────────────
		{
			file:       "models/profile.go", // renamed from models/user.go in Phase 1
			structName: "Profile",
			field:      "\tTimezone   string     `db:\"timezone\"` // IANA tz name, e.g. 'Asia/Dhaka'",
		},
		// ── Contact: short handle + net balance ───────────────────────────────
		// NOTE: contact types may live in models/profile.go (if combined) or in a
		// separate models/contact.go. We try both; whichever exists wins.
		{
			file:       "models/profile.go",
			structName: "Contact",
			field:      "\tHandle     string  `db:\"handle\"`  // short ref used in text parsing, e.g. 'john'",
		},
		{
			file:       "models/profile.go",
			structName: "Contact",
			field:      "\tNetBalance float64 `db:\"net_balance\"` // >0 they owe you; <0 you owe them",
		},
	}

	// Add Wallet Version field if we found the file
	if walletModelFile != "" {
		additions = append(additions, fieldAdd{
			file:       walletModelFile,
			structName: "Wallet",
			field:      "\tVersion    int        `db:\"version\"` // optimistic concurrency lock",
		})
		// Also try the un-renamed name in case Phase 2 hasn't run yet
		additions = append(additions, fieldAdd{
			file:       walletModelFile,
			structName: "Wallet",
			field:      "\tVersion    int        `db:\"version\"` // optimistic concurrency lock",
		})
	}

	for _, a := range additions {
		path := filepath.Join(rootDir, a.file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if *verbose {
				infof("  %s  %s (file not found, skip)\n", dim("SKIP"), a.file)
			}
			continue
		}

		raw, err := os.ReadFile(path)
		must(err, "read "+path)
		original := string(raw)

		// Skip if the field name is already present
		fieldName := strings.Fields(strings.TrimSpace(a.field))[0]
		if strings.Contains(original, fieldName) {
			if *verbose {
				infof("  %s  field '%s' already in %s\n", dim("SKIP"), fieldName, a.file)
			}
			continue
		}

		updated, ok := addFieldToStruct(original, a.structName, a.field)
		if !ok {
			if *verbose {
				infof("  %s  struct '%s' not found in %s\n", dim("SKIP"), a.structName, a.file)
			}
			continue
		}

		if a.importNeeded != "" {
			updated = ensureImport(updated, a.importNeeded)
		}

		logAction("FIELD ADD", fmt.Sprintf("%s :: %s", a.file, strings.TrimSpace(a.field)))
		if !*dryRun {
			must(os.WriteFile(path, []byte(updated), 0o644), "write "+path)
			modifiedFiles = appendUniq(modifiedFiles, path)
		}
	}
}

// findStructInModels searches all .go files under models/ for
// `type <name> struct` and returns the relative path of the first match.
// Multiple candidate names can be supplied (e.g. after and before rename).
func findStructInModels(names ...string) string {
	found := ""
	_ = filepath.Walk(filepath.Join(rootDir, "models"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		src := string(raw)
		for _, name := range names {
			pattern := "type " + name + " struct"
			if strings.Contains(src, pattern) {
				rel, _ := filepath.Rel(rootDir, path)
				found = rel
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

// addFieldToStruct inserts a field line just before the closing } of a named struct.
func addFieldToStruct(src, structName, field string) (string, bool) {
	// Match: type StructName struct { ... }  (multiline)
	pattern := fmt.Sprintf(`(?s)(type\s+%s\s+struct\s*\{)(.*?)(\})`, regexp.QuoteMeta(structName))
	re := regexp.MustCompile(pattern)
	if !re.MatchString(src) {
		return src, false
	}
	result := re.ReplaceAllStringFunc(src, func(match string) string {
		subs := re.FindStringSubmatch(match)
		if len(subs) < 4 {
			return match
		}
		body := subs[2]
		// Ensure field ends with newline
		f := "\n" + field
		if !strings.HasSuffix(f, "\n") {
			f += "\n"
		}
		return subs[1] + body + f + subs[3]
	})
	return result, result != src
}

// ensureImport adds an import path if not already present.
func ensureImport(src, pkg string) string {
	quoted := `"` + pkg + `"`
	if strings.Contains(src, quoted) {
		return src
	}
	// Try to add to existing import block
	re := regexp.MustCompile(`(?s)(import\s*\()([^)]*?)(\))`)
	if re.MatchString(src) {
		return re.ReplaceAllStringFunc(src, func(m string) string {
			subs := re.FindStringSubmatch(m)
			return subs[1] + subs[2] + "\t" + quoted + "\n" + subs[3]
		})
	}
	// Add single import after package declaration
	re2 := regexp.MustCompile(`(?m)^(package\s+\w+\s*)$`)
	return re2.ReplaceAllString(src, "${1}\n\nimport "+quoted)
}

// ════════════════════════════════════════════════════════════════════════════
// PHASE 4 — New source files
// ════════════════════════════════════════════════════════════════════════════

func phaseNewFiles() {
	moduleName := detectModuleName()

	// Each entry: relative path → content function.
	// Paths match the ACTUAL project structure (api/handlers/, modules/transaction/, etc.)
	files := []struct {
		relPath  string
		content  func() string
		skipNote string // non-empty → skip silently with this message
	}{
		{
			// Wizard state store — new directory, safe to create
			relPath: "api/wizard/state.go",
			content: func() string { return newFileWizardState(moduleName) },
		},
		{
			// Health check endpoint — new file under existing pkg/
			relPath: "pkg/health/health.go",
			content: newFileHealth,
		},
		{
			// Message splitting helper — new file under existing pkg/
			relPath: "pkg/telegram/helpers.go",
			content: newFileTelegramHelpers,
		},
		{
			// Undo handler — lives in api/handlers/ alongside wallet.go, user.go, etc.
			relPath: "api/handlers/undo.go",
			content: newFileUndo,
		},
		{
			// Config validation — goes in existing configs/ directory
			relPath: "configs/validate.go",
			content: newFileConfigValidate,
		},
		{
			// Parser tests — the project already has modules/transaction/parser_test.go.
			// We skip creation and print a note pointing to our test cases instead.
			relPath: "modules/transaction/parser_test.go",
			content: func() string { return newFileParserTest(moduleName) },
			skipNote: "modules/transaction/parser_test.go already exists. " +
				"Manually merge the new test cases from the refactor guide into that file.",
		},
		{
			// Cache — the project already has modules/cache/cache.go.
			// We skip to avoid overwriting existing cache logic.
			relPath: "modules/cache/cache.go",
			content: newFileCacheTTL,
			skipNote: "modules/cache/cache.go already exists. " +
				"Review it and add the UserWalletsKey/UserContactsKey helpers if missing.",
		},
	}

	for _, f := range files {
		absPath := filepath.Join(rootDir, f.relPath)

		// Always-skip entries (file already exists in project)
		if f.skipNote != "" {
			infof("  %s  %s\n", yellow("NOTE"), f.skipNote)
			skippedFiles = append(skippedFiles, f.relPath+" (see note above)")
			continue
		}

		// Skip if this particular file already exists on disk
		if _, err := os.Stat(absPath); err == nil {
			if *verbose {
				infof("  %s  %s (already exists, not overwriting)\n", yellow("SKIP"), f.relPath)
			}
			skippedFiles = append(skippedFiles, f.relPath+" (already exists)")
			continue
		}

		logAction("CREATE", f.relPath)
		if !*dryRun {
			must(os.MkdirAll(filepath.Dir(absPath), 0o755), "mkdir for "+f.relPath)
			must(os.WriteFile(absPath, []byte(f.content()), 0o644), "write "+f.relPath)
			createdFiles = append(createdFiles, f.relPath)
		}
	}
}

// ════════════════════════════════════════════════════════════════════════════
// PHASE 5 — Makefile
// ════════════════════════════════════════════════════════════════════════════

func phaseMakefile() {
	path := filepath.Join(rootDir, "Makefile")
	writeOrWarn(path, makefileContent(), "Makefile")
}

// ════════════════════════════════════════════════════════════════════════════
// PHASE 6 — Dockerfile
// ════════════════════════════════════════════════════════════════════════════

func phaseDockerfile() {
	// Primary Dockerfile
	writeOrWarn(filepath.Join(rootDir, "Dockerfile"), dockerfileContent(), "Dockerfile")

	// Dockerfile.in (template variant)
	writeOrWarn(filepath.Join(rootDir, "Dockerfile.in"), dockerfileInContent(), "Dockerfile.in")
}

// ════════════════════════════════════════════════════════════════════════════
// PHASE 7 — CI/CD workflows
// ════════════════════════════════════════════════════════════════════════════

func phaseCICD() {
	wfDir := filepath.Join(rootDir, ".github", "workflows")
	must(os.MkdirAll(wfDir, 0o755), "mkdir .github/workflows")

	writeOrWarn(filepath.Join(wfDir, "ci.yml"), ciWorkflow(), ".github/workflows/ci.yml")
	writeOrWarn(filepath.Join(wfDir, "release.yml"), releaseWorkflow(), ".github/workflows/release.yml")
}

// ════════════════════════════════════════════════════════════════════════════
// PHASE 8 — Config & lint files
// ════════════════════════════════════════════════════════════════════════════

func phaseConfigFiles() {
	writes := map[string]string{
		".env.example":  envExample(),
		".golangci.yml": golangciYML(),
		"CHANGELOG.md":  changelogMD(),
	}
	for rel, content := range writes {
		path := filepath.Join(rootDir, rel)
		if _, err := os.Stat(path); err == nil {
			// Don't overwrite existing .env or CHANGELOG — just warn
			if rel != ".golangci.yml" {
				if *verbose {
					infof("  %s  %s (already exists, not overwriting)\n", yellow("SKIP"), rel)
				}
				continue
			}
		}
		writeOrWarn(path, content, rel)
	}
}

// ════════════════════════════════════════════════════════════════════════════
// FILE CONTENTS
// ════════════════════════════════════════════════════════════════════════════

func newFileWizardState(mod string) string {
	return `package wizard

import (
	"sync"
	"time"
)

// Step identifies the current step in the interactive transaction wizard.
type Step int

const (
	StepType        Step = iota // choose transaction type
	StepCategory                // choose category
	StepSubcategory             // choose subcategory
	StepWallet                  // choose wallet (from / to)
	StepAmount                  // enter amount
	StepDate                    // enter date
	StepNote                    // enter note
	StepConfirm                 // confirm and save
)

// State holds all data collected during an active wizard session.
// It lives server-side so callback_data stays tiny (≤ 64 bytes).
type State struct {
	Step        Step
	TxnType     string
	Category    string
	Subcategory string
	FromWallet  string
	ToWallet    string
	ContactID   int64
	Amount      float64
	Date        string
	Note        string
	ExpiresAt   time.Time
}

// Store is a thread-safe, in-memory wizard state store keyed by Telegram UserID.
type Store struct {
	mu     sync.Mutex
	states map[int64]*State
}

// NewStore creates an empty wizard Store.
func NewStore() *Store {
	return &Store{states: make(map[int64]*State)}
}

// Set stores (or replaces) state for a user with a 10-minute TTL.
func (s *Store) Set(userID int64, state *State) {
	state.ExpiresAt = time.Now().Add(10 * time.Minute)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[userID] = state
}

// Get retrieves state for a user. Returns (nil, false) if not found or expired.
func (s *Store) Get(userID int64) (*State, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.states[userID]
	if !ok {
		return nil, false
	}
	if time.Now().After(st.ExpiresAt) {
		delete(s.states, userID)
		return nil, false
	}
	return st, true
}

// Clear removes wizard state for a user (call on confirm or cancel).
func (s *Store) Clear(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, userID)
}

// PurgeExpired removes all expired entries. Call periodically if needed.
func (s *Store) PurgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, st := range s.states {
		if now.After(st.ExpiresAt) {
			delete(s.states, id)
		}
	}
}
`
}

func newFileCacheTTL() string {
	return `// Package cache provides a simple in-memory TTL cache.
// Use it to avoid hammering the remote database for read-heavy,
// rarely changing data like wallet lists and the category taxonomy.
package cache

import (
	"fmt"
	"sync"
	"time"
)

type entry struct {
	value     interface{}
	expiresAt time.Time
}

// TTLCache is a generic thread-safe in-memory cache with per-entry TTL.
type TTLCache struct {
	mu      sync.RWMutex
	entries map[string]entry
	ttl     time.Duration
}

// New creates a TTLCache whose entries expire after ttl.
func New(ttl time.Duration) *TTLCache {
	c := &TTLCache{
		entries: make(map[string]entry),
		ttl:     ttl,
	}
	// Background purge every 2× TTL
	go func() {
		ticker := time.NewTicker(ttl * 2)
		for range ticker.C {
			c.purge()
		}
	}()
	return c
}

// Set stores value under key, replacing any previous entry.
func (c *TTLCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry{value: value, expiresAt: time.Now().Add(c.ttl)}
}

// Get retrieves a value. Returns (nil, false) if not found or expired.
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

// Invalidate removes a single key immediately.
func (c *TTLCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// UserWalletsKey returns the cache key for a user's wallet list.
func UserWalletsKey(userID int64) string { return fmt.Sprintf("wallets:%d", userID) }

// UserContactsKey returns the cache key for a user's contact list.
func UserContactsKey(userID int64) string { return fmt.Sprintf("contacts:%d", userID) }

// purge removes all expired entries (called in background goroutine).
func (c *TTLCache) purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, k)
		}
	}
}
`
}

func newFileHealth() string {
	// bt holds a single backtick character.
	// Raw string literals cannot contain backticks, so we inject them via concatenation.
	bt := "`"
	return `// Package health exposes a minimal HTTP health-check endpoint.
// Wire it into a lightweight http.ServeMux running on port 8080
// so Railway, Docker HEALTHCHECK, and any monitoring system can
// verify the process is alive and initialised.
//
// In main.go:
//
//	go func() {
//	    mux := http.NewServeMux()
//	    mux.Handle("/health", health.Handler(cmd.Version))
//	    log.Fatal(http.ListenAndServe(":8080", mux))
//	}()
package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type response struct {
	Status    string ` + bt + `json:"status"` + bt + `
	Version   string ` + bt + `json:"version"` + bt + `
	Timestamp string ` + bt + `json:"timestamp"` + bt + `
}

// Handler returns an http.HandlerFunc that responds with JSON {"status":"ok"}.
func Handler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{
			Status:    "ok",
			Version:   version,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}
`
}

func newFileTelegramHelpers() string {
	return `// Package telegram provides shared Telegram bot utilities.
package telegram

import (
	"fmt"
	"strings"
)

const MaxMessageLen = 4000 // safe margin below Telegram's hard 4096-byte limit

// SplitMessage splits a long string into chunks that each fit within
// MaxMessageLen, breaking only on newline boundaries.
// Use this before any bot.Send() call to avoid silent message truncation.
//
// Usage in a handler:
//
//	for _, chunk := range telegram.SplitMessage(text) {
//	    if err := c.Send(chunk, telebot.ModeMarkdown); err != nil {
//	        return err
//	    }
//	}
func SplitMessage(text string) []string {
	if len(text) <= MaxMessageLen {
		return []string{text}
	}
	var chunks []string
	var buf strings.Builder
	for _, line := range strings.Split(text, "\n") {
		// +1 for the newline we are about to add
		if buf.Len()+len(line)+1 > MaxMessageLen {
			if buf.Len() > 0 {
				chunks = append(chunks, buf.String())
				buf.Reset()
			}
		}
		buf.WriteString(line + "\n")
	}
	if buf.Len() > 0 {
		chunks = append(chunks, buf.String())
	}
	return chunks
}

// FormatAmount formats a float64 as a currency string with 2 decimal places.
func FormatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}
`
}

func newFileUndo() string {
	return `package handlers

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

// HandleUndo reverses the most recent active transaction for the calling user.
// It soft-deletes the transaction and reverses any wallet / contact balance
// changes that were applied when it was originally created.
//
// NOTE: Adapt the service call to match your actual TransactionService interface.
// The Undo() method must be implemented separately — see the refactor guide §4.1c.
//
// Register this handler in api/tele.go or bot setup:
//
//	bot.Handle("/undo", h.HandleUndo)
func (h *Handler) HandleUndo(c telebot.Context) error {
	userID := c.Sender().ID

	// Replace this call with however your project passes context to services.
	// Examples:
	//   undone, err := h.txnSvc.Undo(context.Background(), int64(userID))
	//   undone, err := h.txnSvc.Undo(c.Get("ctx").(context.Context), int64(userID))
	undone, err := h.txnSvc.Undo(int64(userID))
	if err != nil {
		return c.Send("❌ Nothing to undo — your transaction history is empty.")
	}

	msg := fmt.Sprintf(
		"✅ *Undone:*\n"+
			"Type: %s\n"+
			"Amount: *%.2f*\n"+
			"Date: %s\n\n"+
			"Wallet balances have been restored.",
		undone.Type, undone.Amount, undone.Date,
	)
	return c.Send(msg, telebot.ModeMarkdown)
}
`
}

func newFileParserTest(mod string) string {
	_ = mod
	return `package modules_test

// Table-driven tests for the natural-language transaction parser.
//
// IMPORTANT: Verify that modules.ParseTransaction and modules.ParsedTxn
// match the actual exported function and struct names in your modules package.
// If the names differ, update the references below accordingly.
//
// Run with:
//
//	go test ./modules/... -v
//
// Add a new row to the 'tests' slice whenever a parsing edge case is
// reported — this prevents regressions as the parser evolves.

import (
	"testing"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/modules"
)

func TestParseTransaction(t *testing.T) {
	t.Parallel()

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, got modules.ParsedTxn)
	}{
		{
			name:  "basic spend with subcategory and wallet",
			input: "spent 1000 for food-rest from dbbl",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "amount",      1000.0,      got.Amount)
				assertEqual(t, "subcategory", "food-rest", got.Subcategory)
				assertEqual(t, "fromWallet",  "dbbl",      got.FromWallet)
			},
		},
		{
			name:  "earn into wallet with explicit date and note",
			input: "earn 5000 to brac on 2024-03-15 note Salary",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "amount",   5000.0,       got.Amount)
				assertEqual(t, "toWallet", "brac",       got.ToWallet)
				assertEqual(t, "date",     "2024-03-15", got.Date)
				assertEqual(t, "note",     "Salary",     got.Note)
			},
		},
		{
			name:  "transfer between wallets",
			input: "transferred 2000 from brac to dbbl on 2024-01-01 note Bill payment",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "fromWallet", "brac",   got.FromWallet)
				assertEqual(t, "toWallet",   "dbbl",   got.ToWallet)
				assertEqual(t, "amount",     2000.0,   got.Amount)
			},
		},
		{
			name:  "lend to contact",
			input: "lend 1000 to john from brac",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "contactHandle", "john", got.ContactHandle)
				assertEqual(t, "fromWallet",    "brac", got.FromWallet)
				assertEqual(t, "amount",        1000.0, got.Amount)
			},
		},
		{
			name:  "borrow from contact",
			input: "borrow 500 from john to cash",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "contactHandle", "john", got.ContactHandle)
				assertEqual(t, "toWallet",      "cash", got.ToWallet)
			},
		},
		{
			name:  "return money to contact",
			input: "return 500 to john from cash",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "contactHandle", "john", got.ContactHandle)
			},
		},
		{
			name:  "recover money from contact",
			input: "recover 800 from john to brac",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "contactHandle", "john", got.ContactHandle)
				assertEqual(t, "toWallet",      "brac", got.ToWallet)
			},
		},
		{
			name:  "quoted note with spaces",
			input: ` + "`" + `spent 500 for food-rest note "Lunch with the team"` + "`" + `,
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "note", "Lunch with the team", got.Note)
			},
		},
		{
			name:  "relative date yesterday",
			input: "spent 200 for trans-taxi yesterday",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "date", yesterday, got.Date)
			},
		},
		{
			name:  "named time morning",
			input: "spent 100 for food-snack at morning",
			check: func(t *testing.T, got modules.ParsedTxn) {
				if got.Time == "" {
					t.Error("expected non-empty time for 'morning'")
				}
			},
		},
		{
			name:  "human date format MMM DD YYYY",
			input: "spent 800 for shop-cloth on Jan 13, 2024",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "date", "2024-01-13", got.Date)
			},
		},
		{
			name:  "DD-MM-YYYY date format",
			input: "spent 300 for health-med on 15-03-2024",
			check: func(t *testing.T, got modules.ParsedTxn) {
				assertEqual(t, "date", "2024-03-15", got.Date)
			},
		},
		// ── Error cases ────────────────────────────────────────────
		{
			name:    "missing amount should error",
			input:   "spent for food-rest from dbbl",
			wantErr: true,
		},
		{
			name:    "negative amount should error",
			input:   "spent -100 for food-rest",
			wantErr: true,
		},
		{
			name:    "unknown action verb should error",
			input:   "frobbled 500 from dbbl",
			wantErr: true,
		},
		{
			name:    "empty input should error",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := modules.ParseTransaction(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseTransaction(%q)\n  got error: %v\n  wantErr:   %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func assertEqual(t *testing.T, field string, want, got interface{}) {
	t.Helper()
	if want != got {
		t.Errorf("field %q:\n  want: %v\n  got:  %v", field, want, got)
	}
}
`
}

func newFileConfigValidate() string {
	return `// Package configs provides configuration loading and validation.
package configs

import (
	"fmt"
	"os"
	"strings"
)

// requiredEnvVars lists environment variables that MUST be set for the bot
// to start. Add new required keys here as they are introduced.
var requiredEnvVars = []string{
	"TELEGRAM_BOT_TOKEN",
	"PARSE_APP_ID",
	"PARSE_REST_API_KEY",
	"PARSE_SERVER_URL",
}

// Validate checks that all required environment variables are present and
// non-empty. Call this at the very start of main() before initialising
// anything else so operators get a clear diagnostic on misconfiguration.
//
//	if err := configs.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func Validate() error {
	var missing []string
	for _, key := range requiredEnvVars {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf(
		"missing required environment variables: %s\n"+
			"Copy .env.example to .env and fill in the values, "+
			"or set them as environment variables.",
		strings.Join(missing, ", "),
	)
}
`
}

// ─── Makefile content ─────────────────────────────────────────────────────────

func makefileContent() string {
	return `# ══════════════════════════════════════════════════════════════════════════════
# Expense Tracker Bot — Makefile
# ══════════════════════════════════════════════════════════════════════════════

# ── Variables ──────────────────────────────────────────────────────────────────
BINARY      := expense-tracker
MODULE      := $(shell go list -m 2>/dev/null || echo github.com/masudur-rahman/expense-tracker-bot)
REGISTRY    ?= docker.io
IMAGE_REPO  ?= masudurrahman/expense-tracker-bot
PLATFORMS   ?= linux/amd64,linux/arm64
BOT_ENV     ?= dev

# Build metadata (embedded into binary via ldflags)
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -ldflags "-s -w \
  -X $(MODULE)/cmd.Version=$(VERSION) \
  -X $(MODULE)/cmd.GitCommit=$(GIT_COMMIT) \
  -X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)"

TAG         ?= $(VERSION)

# Enable BuildKit for all docker commands
export DOCKER_BUILDKIT := 1

# ── Build ──────────────────────────────────────────────────────────────────────
.PHONY: build
build: ## Build binary for the current platform
	@mkdir -p bin
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY) .

.PHONY: build-linux
build-linux: ## Build binary for linux/amd64
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 .

.PHONY: cross-build
cross-build: ## Cross-compile for all target platforms → dist/
	@mkdir -p dist
	@for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
	  OS=$$(echo $$platform | cut -d/ -f1); \
	  ARCH=$$(echo $$platform | cut -d/ -f2); \
	  EXT=$$([ "$$OS" = "windows" ] && echo ".exe" || echo ""); \
	  OUTPUT=dist/$(BINARY)-$$OS-$$ARCH$$EXT; \
	  echo "  → $$OUTPUT"; \
	  CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH \
	    go build $(LDFLAGS) -o $$OUTPUT . || exit 1; \
	done
	@echo "Cross-build complete. Binaries in dist/"

# ── Test & Quality ─────────────────────────────────────────────────────────────
.PHONY: test
test: ## Run all tests with race detector and coverage
	go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep -E "^total:" | awk '{print "Coverage: " $$3}'

.PHONY: test-short
test-short: ## Run tests without long-running cases
	go test -short ./...

.PHONY: coverage-html
coverage-html: test ## Open coverage report in browser
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./... --timeout=5m

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: vulncheck
vulncheck: ## Run govulncheck for known vulnerabilities
	@which govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

.PHONY: check
check: vet lint test ## Run all quality checks (vet + lint + test)

# ── Run ────────────────────────────────────────────────────────────────────────
.PHONY: run
run: build ## Build and run locally
	BOT_ENV=$(BOT_ENV) ./bin/$(BINARY)

.PHONY: run-polling
run-polling: build ## Run in polling mode (best for local dev)
	BOT_ENV=$(BOT_ENV) BOT_MODE=polling ./bin/$(BINARY)

# ── Docker — single arch (fast, for local testing) ────────────────────────────
.PHONY: docker-build
docker-build: ## Build Docker image for current platform
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg BUILD_DATE=$(BUILD_DATE) \
	  --build-arg GIT_COMMIT=$(GIT_COMMIT) \
	  -t $(REGISTRY)/$(IMAGE_REPO):$(TAG) \
	  -t $(REGISTRY)/$(IMAGE_REPO):latest \
	  -f Dockerfile .

.PHONY: docker-push
docker-push: ## Push single-arch image to registry
	docker push $(REGISTRY)/$(IMAGE_REPO):$(TAG)
	docker push $(REGISTRY)/$(IMAGE_REPO):latest

# ── Docker — multi-arch (for releases) ────────────────────────────────────────
.PHONY: docker-setup-buildx
docker-setup-buildx: ## Create/activate the multiarch buildx builder
	@docker buildx inspect multiarch >/dev/null 2>&1 \
	  && docker buildx use multiarch \
	  || docker buildx create --name multiarch --driver docker-container --use
	docker buildx inspect --bootstrap

.PHONY: docker-buildx
docker-buildx: docker-setup-buildx ## Build & push multi-arch image (amd64 + arm64)
	docker buildx build \
	  --platform $(PLATFORMS) \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg BUILD_DATE=$(BUILD_DATE) \
	  --build-arg GIT_COMMIT=$(GIT_COMMIT) \
	  --tag $(REGISTRY)/$(IMAGE_REPO):$(TAG) \
	  --tag $(REGISTRY)/$(IMAGE_REPO):latest \
	  --push \
	  -f Dockerfile .
	@echo "Multi-arch image pushed: $(REGISTRY)/$(IMAGE_REPO):$(TAG)"

# ── Release ────────────────────────────────────────────────────────────────────
.PHONY: release
release: check cross-build docker-buildx ## Full release: test + cross-compile + multi-arch Docker push
	@echo "Release $(VERSION) complete ✓"

# ── Utilities ──────────────────────────────────────────────────────────────────
.PHONY: tidy
tidy: ## Tidy and verify Go modules
	go mod tidy
	go mod verify

.PHONY: clean
clean: ## Remove build artefacts
	rm -rf bin/ dist/ coverage.out

.PHONY: version
version: ## Print the current version
	@echo $(VERSION)

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | sort \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
`
}

// ─── Dockerfile content ───────────────────────────────────────────────────────

func dockerfileContent() string {
	return `# syntax=docker/dockerfile:1
# ══════════════════════════════════════════════════════════════════════════════
# Stage 1 · Build
# Uses BuildKit cache mounts so Go modules and the build cache persist
# between CI runs, dramatically reducing build time.
# ══════════════════════════════════════════════════════════════════════════════
FROM golang:1.22-alpine AS builder

ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=none

WORKDIR /app

# Download dependencies first — this layer is only rebuilt when go.mod/go.sum change
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Build the binary — source changes only rebuild from here
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build \
      -ldflags "-s -w \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.Version=${VERSION} \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.BuildDate=${BUILD_DATE} \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.GitCommit=${GIT_COMMIT}" \
      -o /expense-tracker .

# ══════════════════════════════════════════════════════════════════════════════
# Stage 2 · Runtime
# distroless/static:nonroot — no shell, no package manager, no root.
# Final image is ~5-8 MB.  Attack surface is minimal.
# ══════════════════════════════════════════════════════════════════════════════
FROM gcr.io/distroless/static:nonroot

# The :nonroot tag sets USER 65532 (nonroot) by default, but be explicit
USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /expense-tracker /expense-tracker

# Health check port (requires a /health HTTP endpoint — see pkg/health/health.go)
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD ["/expense-tracker", "health"]

ENTRYPOINT ["/expense-tracker"]
`
}

func dockerfileInContent() string {
	return `# syntax=docker/dockerfile:1
# Template Dockerfile — Makefile substitutes @VERSION@, @BUILD_DATE@, @GIT_COMMIT@
# before passing to docker build.  Keep in sync with Dockerfile.

FROM golang:1.22-alpine AS builder

ARG VERSION=@VERSION@
ARG BUILD_DATE=@BUILD_DATE@
ARG GIT_COMMIT=@GIT_COMMIT@

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build \
      -ldflags "-s -w \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.Version=${VERSION} \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.BuildDate=${BUILD_DATE} \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.GitCommit=${GIT_COMMIT}" \
      -o /expense-tracker .

FROM gcr.io/distroless/static:nonroot
USER nonroot:nonroot
COPY --from=builder --chown=nonroot:nonroot /expense-tracker /expense-tracker
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD ["/expense-tracker", "health"]
ENTRYPOINT ["/expense-tracker"]
`
}

// ─── CI/CD workflow content ────────────────────────────────────────────────────

func ciWorkflow() string {
	return `name: CI

on:
  push:
    branches: [ main, natural ]
  pull_request:
    branches: [ main, natural ]

jobs:
  # ────────────────────────────────────────────────────────────────
  # Job 1 · Tests
  # ────────────────────────────────────────────────────────────────
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Cache Go modules & build cache
        uses: actions/cache@v4
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: ${{ runner.os }}-go-

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: coverage.out
          fail_ci_if_error: false

  # ────────────────────────────────────────────────────────────────
  # Job 2 · Lint
  # ────────────────────────────────────────────────────────────────
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m

  # ────────────────────────────────────────────────────────────────
  # Job 3 · Vulnerability scan
  # ────────────────────────────────────────────────────────────────
  vuln:
    name: Vuln Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

  # ────────────────────────────────────────────────────────────────
  # Job 4 · Build (depends on test + lint passing)
  # ────────────────────────────────────────────────────────────────
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [ test, lint ]
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # needed for git describe

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Cache Go modules & build cache
        uses: actions/cache@v4
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: ${{ runner.os }}-go-

      - name: Build
        run: make build
`
}

func releaseWorkflow() string {
	return `name: Release

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest
    permissions:
      contents: write   # create GitHub release
      packages: write   # push to GHCR (if used)

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # needed for git describe

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Cache Go modules & build cache
        uses: actions/cache@v4
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: ${{ runner.os }}-go-

      - name: Cross-compile binaries
        run: make cross-build

      - name: Set up QEMU (for ARM emulation)
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build & push multi-arch image
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: |
            docker.io/masudurrahman/expense-tracker-bot:${{ github.ref_name }}
            docker.io/masudurrahman/expense-tracker-bot:latest
          cache-from: type=gha
          cache-to:   type=gha,mode=max
          build-args: |
            VERSION=${{ github.ref_name }}
            BUILD_DATE=${{ github.event.head_commit.timestamp }}
            GIT_COMMIT=${{ github.sha }}

      - name: Scan image with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: docker.io/masudurrahman/expense-tracker-bot:${{ github.ref_name }}
          format: table
          exit-code: '1'
          severity: 'CRITICAL,HIGH'

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
          generate_release_notes: true
`
}

// ─── Config file content ──────────────────────────────────────────────────────

func envExample() string {
	return `# ══════════════════════════════════════════════════════════════════════════════
# Expense Tracker Bot — Environment Variables
# Copy this file to .env and fill in your values.
# Never commit a filled-in .env to version control.
# ══════════════════════════════════════════════════════════════════════════════

# ── Telegram ───────────────────────────────────────────────────────────────────
# Get this from @BotFather on Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_here

# 'polling' for local dev (no HTTPS needed), 'webhook' for production
BOT_MODE=polling

# Required only when BOT_MODE=webhook — must be an HTTPS URL
WEBHOOK_URL=https://your-domain.com/telegram/webhook

# Your personal Telegram numeric user ID
# Find it by messaging @userinfobot on Telegram
# Used to restrict /backup and other admin-only commands
OWNER_TELEGRAM_ID=123456789

# ── Database: Back4App (default) ───────────────────────────────────────────────
PARSE_APP_ID=your_back4app_application_id
PARSE_REST_API_KEY=your_back4app_rest_api_key
PARSE_SERVER_URL=https://parseapi.back4app.com

# ── Database: Supabase (alternative) ─────────────────────────────────────────
# Uncomment to use Supabase instead of Back4App
# SUPABASE_URL=https://your-project-ref.supabase.co
# SUPABASE_ANON_KEY=your_supabase_anon_key

# ── Database: Direct PostgreSQL (alternative / Railway add-on) ────────────────
# Uncomment to use a direct Postgres connection
# DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require

# ── Logging ────────────────────────────────────────────────────────────────────
# debug | info | warn | error   (default: info)
LOG_LEVEL=info

# ── Application ────────────────────────────────────────────────────────────────
# dev | prod   (selects config profile and adjusts defaults)
BOT_ENV=dev
`
}

func golangciYML() string {
	return `# golangci-lint configuration
# Docs: https://golangci-lint.run/usage/configuration/

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
  govet:
    enable-all: true
  staticcheck:
    checks: ["all"]
  misspell:
    locale: US
  gosec:
    excludes:
      - G304  # file path from variable (acceptable in our config loader)

linters:
  enable:
    - errcheck       # ensure errors are always checked
    - govet          # go vet analysis
    - gosimple       # simplification suggestions
    - staticcheck    # advanced static analysis
    - ineffassign    # detect ineffectual assignments
    - unused         # detect unused code
    - gofmt          # formatting check
    - gosec          # security issues
    - misspell       # spelling errors in comments/strings
    - bodyclose      # HTTP response body must be closed
    - noctx          # HTTP requests should use context
    - contextcheck   # context propagation checks
    - nilerr         # return nil err instead of nil, nil
    - errorlint      # error wrapping correctness

  disable:
    - depguard       # we manage deps via go mod

run:
  timeout: 5m
  skip-dirs:
    - vendor

issues:
  exclude-rules:
    # Relax some rules in test files
    - path: _test\.go
      linters:
        - gosec
        - errcheck
        - bodyclose
`
}

func changelogMD() string {
	now := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`# Changelog

All notable changes to this project are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] — natural branch

### Added
- **Naming refactor**: Wallet → Wallet, Contact/Contact → Contact, User (bot owner) → Profile
- **Soft-delete**: Transaction.DeletedAt field; /undo command to reverse the most recent transaction
- **Optimistic locking**: Wallet.Version field to prevent concurrent balance race conditions
- **Wizard state store**: Server-side state for /newtxn flow (fixes 64-byte callback_data limit)
- **Message splitting**: SplitMessage() helper to stay within Telegram's 4096-byte message limit
- **Health endpoint**: pkg/health HTTP handler for Railway / Docker HEALTHCHECK
- **TTL cache**: pkg/cache for wallet lists and category taxonomy (reduces Back4App round-trips)
- **Config validation**: configs.Validate() fails fast on startup if required env vars are missing
- **.env.example**: Template listing all required/optional environment variables
- **.golangci.yml**: Comprehensive lint configuration
- **Multi-arch Docker**: Makefile docker-buildx target and CI release.yml using docker/build-push-action
- **Cross-compilation**: Makefile cross-build target for linux, darwin, windows (amd64 + arm64)
- **ldflags**: VERSION, BUILD_DATE, GIT_COMMIT embedded into binary at build time
- **BuildKit**: DOCKER_BUILDKIT=1 enabled globally in Makefile; cache mounts in Dockerfile
- **Distroless runtime**: Dockerfile now uses gcr.io/distroless/static:nonroot (from Alpine)
- **Non-root container**: USER nonroot:nonroot in Dockerfile
- **CI pipeline**: .github/workflows/ci.yml with test, lint, vuln scan, build jobs
- **Release pipeline**: .github/workflows/release.yml with multi-arch Docker + GitHub Release
- **Parser tests**: modules/parser_test.go with table-driven cases
- **CHANGELOG.md**: This file

### Changed
- /users command renamed to /contacts
- Menu labels: "Wallet" → "Wallet", "Contacts" → "Contact"
- Contact.Balance renamed to Contact.NetBalance (positive = they owe you)
- Contact struct gains Handle field (short name for text parsing)
- Profile struct gains Timezone field

### Natural branch parser improvements
- Wider action-verb vocabulary (paid, received, repaid, collected, ...)
- Quoted note fields: note "Lunch with team"
- Relative date expressions: yesterday, last monday, -3d
- Fuzzy wallet name matching (partial names resolve if unambiguous)
- Descriptive error messages on parse failure

---

## [v1.0.0] — %s

- Initial public release
- Telegram bot for tracking daily transactions
- Interactive /newtxn flow and natural language text parsing
- 13-category / 80+ subcategory taxonomy
- PDF report generation
- Back4App database backend via styx library
- Railway deployment support
`, now)
}

// ════════════════════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════════════════════

func walkGoFiles(fn func(string) error) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip vendor/, hidden dirs, and non-.go files
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || (len(name) > 1 && name[0] == '.') {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		return fn(path)
	})
}

func verifyProject() error {
	gomod := filepath.Join(rootDir, "go.mod")
	data, err := os.ReadFile(gomod)
	if err != nil {
		return fmt.Errorf("go.mod not found: %w", err)
	}
	if !strings.Contains(string(data), "expense-tracker-bot") {
		return fmt.Errorf("go.mod does not mention expense-tracker-bot — wrong directory?")
	}
	return nil
}

func detectModuleName() string {
	data, err := os.ReadFile(filepath.Join(rootDir, "go.mod"))
	if err != nil {
		return "github.com/masudur-rahman/expense-tracker-bot"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "github.com/masudur-rahman/expense-tracker-bot"
}

func createBackup() error {
	ts := time.Now().Format("20060102-150405")
	dest := fmt.Sprintf(".refactor-backup-%s", ts)
	infof("Creating backup → %s\n", dest)
	return copyDir(rootDir, dest, func(path string) bool {
		base := filepath.Base(path)
		return base == "vendor" || base == ".git" || strings.HasPrefix(base, ".refactor-backup")
	})
}

func copyDir(src, dst string, skip func(string) bool) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if skip != nil && skip(filepath.Base(path)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, mode)
}

func writeOrWarn(path, content, label string) {
	logAction("WRITE", label)
	if !*dryRun {
		must(os.MkdirAll(filepath.Dir(path), 0o755), "mkdir "+filepath.Dir(path))
		must(os.WriteFile(path, []byte(content), 0o644), "write "+label)
		createdFiles = appendUniq(createdFiles, label)
	}
}

func runPhase(name string, enabled bool, fn func()) {
	if !enabled {
		infof("%s  %s\n", dim("[SKIP]"), dim(name))
		return
	}
	fmt.Printf("\n%s %s\n", bold(blue("▶")), bold(name))
	fmt.Printf("%s\n", strings.Repeat("─", 70))
	fn()
}

func logAction(kind, target string) {
	col := cyan
	switch kind {
	case "RENAME":
		col = yellow
	case "CREATE", "WRITE":
		col = green
	case "MODIFY":
		col = blue
	case "FIELD ADD":
		col = cyan
	}
	prefix := fmt.Sprintf("%-12s", kind)
	fmt.Printf("  %s %s\n", col(prefix), target)
}

func infof(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, red("ERROR: ")+format+"\n", args...)
	os.Exit(1)
}

func must(err error, context string) {
	if err != nil {
		fatalf("%s: %v", context, err)
	}
}

func parseSkip(s string) map[string]bool {
	m := map[string]bool{}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			m[p] = true
		}
	}
	return m
}

func appendUniq(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

// ─── Summary & manual steps ───────────────────────────────────────────────────

func printSummary() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 70))
	fmt.Printf("%s\n", bold("  SUMMARY"))
	fmt.Printf("%s\n\n", strings.Repeat("═", 70))

	if *dryRun {
		fmt.Printf("  %s  No files were changed.\n\n", yellow("[DRY-RUN]"))
	}

	printList(green("  ✓ Renamed"), renamedFiles)
	printList(blue("  ✓ Modified"), modifiedFiles)
	printList(green("  ✓ Created"), createdFiles)
	printList(dim("  ~ Skipped"), skippedFiles)

	if len(warnings) > 0 {
		fmt.Printf("\n%s\n", yellow(bold("  ⚠  WARNINGS")))
		for i, w := range warnings {
			fmt.Printf("  %d. %s\n\n", i+1, yellow(w))
		}
	}

	// Always-present manual steps
	manualSteps = append(manualSteps,
		"Add TransactionService.Undo() implementation: soft-delete the last transaction "+
			"and call revertWalletDelta() to reverse balance changes. See refactor guide §4.1c.",
		"Add TransactionService.Create() compensating-action pattern: if wallet or contact "+
			"balance update fails, soft-delete the newly saved transaction. See refactor guide §4.1c.",
		"Implement ContactRepo and WalletRepo UpdateBalance with optimistic lock (version field). "+
			"See refactor guide §10.1.",
		"Wire the wizard.Store into your /newtxn handler so callback_data only carries a step "+
			"identifier. See refactor guide §5.5.",
		"Wire SendSplit() into /expense, /summary, /allsummary handlers. See pkg/telegram/helpers.go.",
		"Register /contacts and /undo commands in your bot setup (bot.Handle).",
		"Add 'go run refactor.go health' sub-command so the Dockerfile HEALTHCHECK works. "+
			"See pkg/health/health.go.",
		"If models/user.go was renamed to profile.go and contained Contact data too, manually "+
			"extract the Contact struct into models/contact.go.",
		"If using wkhtmltopdf for PDF generation, migrate to a pure-Go library (gofpdf or gopdf) "+
			"so the distroless image can be used without extra system packages.",
		"Add DOCKERHUB_USERNAME and DOCKERHUB_TOKEN secrets to your GitHub repository settings "+
			"for the release workflow to push Docker images.",
		"Enable branch protection on 'natural' and 'main': require test + lint + build jobs "+
			"to pass before merging.",
	)

	fmt.Printf("\n%s\n", bold(red("  ✎  MANUAL STEPS REQUIRED")))
	fmt.Println("  These changes require human judgment and cannot be automated:")
	for i, step := range manualSteps {
		fmt.Printf("\n  %s %s\n", bold(fmt.Sprintf("%d.", i+1)), step)
	}

	fmt.Printf("\n%s\n\n", strings.Repeat("═", 70))
	fmt.Printf("  Run %s to check for compilation errors.\n", cyan("go build ./..."))
	fmt.Printf("  Run %s to run all tests.\n", cyan("go test ./..."))
	fmt.Printf("  Run %s for the full quality check.\n\n", cyan("make check"))
}

func printList(label string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Printf("%s (%d)\n", label, len(items))
	for _, item := range items {
		fmt.Printf("    %s %s\n", dim("•"), item)
	}
	fmt.Println()
}

// ─── Banner ───────────────────────────────────────────────────────────────────

func printBanner() {
	fmt.Println()
	fmt.Println(bold(cyan("  ╔══════════════════════════════════════════════════════════════╗")))
	fmt.Println(bold(cyan("  ║       Expense Tracker Bot — Automated Refactor Script        ║")))
	fmt.Println(bold(cyan("  ║                 Branch: natural                              ║")))
	fmt.Println(bold(cyan("  ╚══════════════════════════════════════════════════════════════╝")))
	fmt.Println()
	fmt.Printf("  %s   %s\n", bold("Dry-run:"), yesNo(*dryRun))
	fmt.Printf("  %s  %s\n", bold("Verbose:"), yesNo(*verbose))
	fmt.Printf("  %s     %s\n", bold("Root:"), rootDir)
	fmt.Printf("  %s  %s\n\n", bold("Backup:"), dim("auto-created before changes"))
}

func yesNo(b bool) string {
	if b {
		return green("yes")
	}
	return dim("no")
}
