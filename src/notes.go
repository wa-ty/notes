package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Path   string
	Editor string
	Sync   string
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}

func main() {
	var config Config
	data, _ := os.ReadFile(os.Getenv("NOTESCONFIGFILE"))
	err := yaml.Unmarshal([]byte(data), &config)
	check(err)
	config.Path = place_home(config.Path)
	folder_name := time.Now().Format("01_02_2006")

	// Flag parsing

	sync := flag.Bool("sync", false, "sync directory")
	code := flag.Bool("code", false, "using vscode")
	vim := flag.Bool("vim", false, "using vim")
	list := flag.Bool("list", false, "list files in this directory")
	choose := flag.Bool("choose", false, "choose file in directory")
	delete := flag.Bool("delete", false, "delete chosen file")
	cd := flag.Bool("cd", false, "shell into the directory")

	flag.Parse()
	flags := flag.Args()

	// Conf check

	if !exists(config.Path) {
		fmt.Println(sync, "No directory at path", config.Path, "(conf.yml)")
		os.Exit(1)
	}

	// File name

	file_name := "default"
	if len(flags) == 1 {
		file_name = flags[0]
	} else if len(flags) > 1 {
		fmt.Println("Invalid arguments")
		os.Exit(1)
	}

	// Editor

	editor := config.Editor
	if len(editor) == 0 {
		editor = "vim"
	}

	if *code {
		editor = "code"
	} else if *vim {
		editor = "vim"
	}

	// folder path

	folder_path := filepath.Join(config.Path, folder_name)

	if !exists(folder_path) {
		err := os.Mkdir(folder_path, 0755)
		check(err)
	}

	// file path

	file_path := filepath.Join(folder_path, file_name)
	if *choose {
		file_path = choose_file(folder_path)
	}

	// Go

	if *sync {
		sync_folder(folder_path)
	} else if *delete {
		delete_file(file_path)
	} else if *cd {
		change_dir(folder_path)
	} else if *list {
		list_files(folder_path, false)
	} else {
		open_file(file_path, editor)
	}
}

func sync_folder(folder_path string) {
	fmt.Println("syncing folder...")
}

func delete_file(file_path string) {
	if !exists(file_path) {
		fmt.Println(file_path + " does not exist")
		return
	}
	var choice string
	fmt.Println("(y/n) delete " + file_path + "?")
	fmt.Scanln(&choice)
	if strings.ToLower(choice) == "y" {
		err := os.Remove(file_path)
		check(err)
	}
}

func change_dir(folder_path string) {
	fmt.Println("now in child process, exit to return")
	cmd := exec.Command("zsh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Dir = folder_path
	err := cmd.Run()
	check(err)
}

func choose_file(folder_path string) string {
	files := list_files(folder_path, true)
	if len(files) <= 0 {
		return ""
	}
	var choice int
	var string_choice string
	for {
		fmt.Scanln(&string_choice)
		choice, err := strconv.Atoi(string_choice)
		if in(string_choice, files) {
			break
		} else if err != nil {
			continue
		} else if choice > 0 && choice <= len(files) {
			break
		}
	}
	return filepath.Join(folder_path, files[choice])
}

func open_file(file_path string, using string) {
	cmd := exec.Command(using, file_path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	check(err)
}

func list_files(folder_path string, numbers bool) []string {
	var file_names []string
	files, err := ioutil.ReadDir(folder_path)
	check(err)
	pre := "-"
	for i, f := range files {
		if numbers {
			pre = fmt.Sprintf("%d", (i+1)) + ")"
		}
		edit_time := "(" + f.ModTime().Format("Jan _2 3:04PM") + ")"
		fmt.Println(pre, f.Name(), "\t", edit_time)
		file_names = append(file_names, f.Name())
	}
	return file_names
}
