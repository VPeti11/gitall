package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var (
	addRepo     = flag.String("addrepo", "", "Add repo path to Database")
	delRepo     = flag.String("delrepo", "", "Delete repo path from Database")
	reinit      = flag.Bool("reinit", false, "Reinitialize the Database")
	command     = flag.Bool("command", false, "Run a git command on all repos")
	listRepos   = flag.Bool("listrepos", false, "List all repos in the Database")
	listLastOps = flag.Bool("listops", false, "List last 50 operations")

	excludeList = flag.String("exclude", "", "Comma-separated list of repo paths to exclude")
	dbFile      = flag.String("db", "", "Specify alternate Database file")
)

func defaultPaths() (string, string, string) {
	usr, _ := user.Current()
	base := filepath.Join(usr.HomeDir, ".gitall.db")
	return base, base + ".sha256", base + ".log"
}

func resolvePaths() (string, string, string) {
	if *dbFile != "" {
		base := *dbFile
		return base, base + ".sha256", base + ".log"
	}
	return defaultPaths()
}

func checkGitInstalled() error {
	_, err := exec.LookPath("git")
	return err
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return []string{}, nil
	}
	defer file.Close()
	var lines []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}

func computeSHA(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeSHA(shaPath, dbPath string) error {
	sum, err := computeSHA(dbPath)
	if err != nil {
		return err
	}
	return os.WriteFile(shaPath, []byte(sum), 0644)
}

func verifySHA(shaPath, dbPath string) error {
	expected, err := os.ReadFile(shaPath)
	if err != nil {
		return fmt.Errorf("SHA file missing")
	}
	actual, err := computeSHA(dbPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(expected)) != actual {
		return fmt.Errorf("SHA mismatch")
	}
	return nil
}

func appendLog(logPath, line string) {
	lines, _ := readLines(logPath)
	lines = append([]string{line}, lines...)
	if len(lines) > 50 {
		lines = lines[:50]
	}
	_ = writeLines(logPath, lines)
}

func addRepoToDB(path, dbPath, shaPath string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !isGitRepo(abs) {
		return fmt.Errorf("not a git repo: %s", abs)
	}
	lines, _ := readLines(dbPath)
	for _, l := range lines {
		if l == abs {
			return fmt.Errorf("already in db")
		}
	}
	lines = append(lines, abs)
	if err := writeLines(dbPath, lines); err != nil {
		return err
	}
	return writeSHA(shaPath, dbPath)
}

func deleteRepoFromDB(path, dbPath, shaPath string) error {
	abs, _ := filepath.Abs(path)
	lines, _ := readLines(dbPath)
	var out []string
	for _, l := range lines {
		if l != abs {
			out = append(out, l)
		}
	}
	if err := writeLines(dbPath, out); err != nil {
		return err
	}
	return writeSHA(shaPath, dbPath)
}

func reinitDB(dbPath, shaPath, logPath string) error {
	_ = os.RemoveAll(dbPath)
	_ = os.RemoveAll(shaPath)
	_ = os.RemoveAll(logPath)
	_ = writeLines(dbPath, []string{})
	return writeSHA(shaPath, dbPath)
}

func runGitCommand(dbPath, shaPath, logPath string, cmdArgs []string, exclude map[string]bool) error {
	if err := verifySHA(shaPath, dbPath); err != nil {
		return err
	}
	repos, err := readLines(dbPath)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if exclude[repo] {
			continue
		}
		if !isGitRepo(repo) {
			fmt.Fprintf(os.Stderr, "Skipping: %s\n", repo)
			continue
		}
		fmt.Printf("Running in: %s\n", repo)
		cmd := exec.Command("git", cmdArgs...)
		cmd.Dir = repo
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error in %s: %v\n", repo, err)
		} else {
			appendLog(logPath, fmt.Sprintf("%s: git %s", repo, strings.Join(cmdArgs, " ")))
		}
	}
	return nil
}

func main() {
	flag.Parse()
	args := flag.Args()
	dbPath, shaPath, logPath := resolvePaths()

	if err := checkGitInstalled(); err != nil {
		fmt.Println("Error: git is not installed or not in PATH.")
		os.Exit(1)
	}

	excludeMap := map[string]bool{}
	if *excludeList != "" {
		paths := strings.Split(*excludeList, ",")
		for _, p := range paths {
			abs, _ := filepath.Abs(p)
			excludeMap[abs] = true
		}
	}

	if len(os.Args) == 1 ||
		(*addRepo == "" &&
			*delRepo == "" &&
			!*reinit &&
			!*command &&
			!*listRepos &&
			!*listLastOps) {

		fmt.Println("gitall - multi-repo git helper")
		fmt.Println("Usage:")
		fmt.Println("  -db <file>               Use custom Database file (default: ~/.gitall.db)")
		fmt.Println("  -addrepo <path>          Add a Git repo to the Database")
		fmt.Println("  -delrepo <path>          Remove a Git repo from the Database")
		fmt.Println("  -reinit                  Reinitialize the Database and SHA/log")
		fmt.Println("  -listrepos               List all repositories in the Database")
		fmt.Println("  -listops       			Show last 50 operations executed")
		fmt.Println("  -exclude <paths>         Comma separated list of repo paths to exclude from -command")
		fmt.Println("  -command <git args>      Run a git command in all repos in the Database")
		os.Exit(0)
	}

	switch {
	case *addRepo != "":
		if err := addRepoToDB(*addRepo, dbPath, shaPath); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Repo added")

	case *delRepo != "":
		if err := deleteRepoFromDB(*delRepo, dbPath, shaPath); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Repo deleted")

	case *reinit:
		if err := reinitDB(dbPath, shaPath, logPath); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Database reset")

	case *listRepos:
		repos, _ := readLines(dbPath)
		for _, r := range repos {
			fmt.Println(r)
		}

	case *listLastOps:
		logs, _ := readLines(logPath)
		for _, l := range logs {
			fmt.Println(l)
		}

	case *command:
		if len(args) == 0 {
			fmt.Println("Error: No git command specified after -command")
			os.Exit(1)
		}
		if err := runGitCommand(dbPath, shaPath, logPath, args, excludeMap); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}
}
