package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/BurntSushi/toml"
	"gopkg.in/src-d/go-git.v4"
)

type config struct {
	LibsDir string
}

type lib struct {
	URL        string
	Hash       string
	Name       string
	LastUpdate time.Time
}

type libs struct {
	Libs []lib
}

const fnameCfg = "Lovecfg"
const fnameLibs = "Lovelibs"

var libsConfig config

func check(err error) {
	if err != nil {
		//fmt.Println(err)
		//os.Exit(1)
		panic(err)
	}
}

func initialize() {
	type step int
	temp := make(map[string]interface{})
	const (
		CheckArguments  step = 1
		CheckFileExists step = 2
		FileExists      step = 3
		Overwrite       step = 4
		WriteToFile     step = 5
		End             step = 0
	)
	nextStep := CheckArguments

	for nextStep != End {
		switch nextStep {
		case CheckArguments:
			if len(os.Args) < 3 {
				fmt.Println("Número de argumentos incorreto.")
				os.Exit(1)
			}
			temp["libsdir"] = os.Args[2]
			nextStep = CheckFileExists
			break

		case CheckFileExists:
			expath, err := exPath()
			check(err)
			exists, err := pathExists(filepath.Join(expath, fmt.Sprintf("%s.%s", fnameCfg, "toml")))
			check(err)
			if exists {
				nextStep = Overwrite
			} else {
				nextStep = WriteToFile
			}
			break

		case Overwrite:
			reader := bufio.NewReader(os.Stdin)
			result := ""
			for result != "y" && result != "n" {
				fmt.Println("Já exite um arquivo de configuração iniciado aqui. Deseja sobrescrevê-lo? [y/n]")
				r, err := reader.ReadString('\n')
				result = strings.ToLower(r)[:len(r)-1]
				check(err)
			}
			if result == "n" {
				nextStep = End
			} else {
				nextStep = WriteToFile
			}
			break

		case WriteToFile:
			cfg := config{
				LibsDir: os.Args[2],
			}

			file, err := os.OpenFile(fmt.Sprintf("%s.%s", fnameCfg, "toml"), os.O_WRONLY|os.O_CREATE, 0644)
			check(err)

			err = toml.NewEncoder(file).Encode(cfg)
			check(err)

			file.Close()

			fmt.Println("Configurações escritas.")

			nextStep = End
			break

		case End:
			os.Exit(0)
		}
	}
}

func addLib() {
	type step int
	temp := make(map[string]interface{})
	const (
		CheckArguments  step = 1
		CheckFileExists step = 2
		CreateFile      step = 3
		CheckRepoExists step = 4
		RepoExists      step = 5
		PullRepo        step = 6
		CreateDir       step = 7
		CloneToDir      step = 8
		CheckOutHash    step = 9
		WriteToFile     step = 10
		End             step = 0
	)
	nextStep := CheckArguments

	for nextStep != End {
		switch nextStep {
		case CheckArguments:
			if len(os.Args) < 4 {
				fmt.Println("Número de argumentos incorreto.")
				os.Exit(1)
			}
			temp["URL"] = os.Args[2]
			temp["Name"] = os.Args[3]
			if len(os.Args) == 5 {
				temp["Hash"] = os.Args[4]
			}
			nextStep = CheckFileExists
			break

		case CheckFileExists:
			expath, err := exPath()
			check(err)
			exists, err := pathExists(filepath.Join(expath, fmt.Sprintf("%s.%s", fnameLibs, "toml")))
			check(err)
			if exists {
				nextStep = CheckRepoExists
			} else {
				temp["createfile"] = fmt.Sprintf("%s.%s", fnameLibs, "toml")
				nextStep = CreateFile
				temp["libs"] = libs{}
			}
			break

		case CreateFile:
			file, err := os.Create(temp["createfile"].(string))
			check(err)
			file.Close()
			nextStep = CreateDir
			break

		case CheckRepoExists:
			llibs := libs{}
			_, err := toml.DecodeFile(fmt.Sprintf("%s.%s", fnameLibs, "toml"), &llibs)
			check(err)
			nextStep = CreateDir
			for i := 0; i < len(llibs.Libs); i++ {
				if llibs.Libs[i].URL == temp["URL"].(string) {
					nextStep = RepoExists
					temp["repo"] = i
					break
				}
			}
			temp["libs"] = llibs
			break

		case RepoExists:
			t := temp["libs"].(libs)
			i := temp["repo"].(int)
			dhash := compareHash(t.Libs[i].Hash, temp["Hash"].(string))
			if !dhash {
				nextStep = PullRepo
				break
			}

			reader := bufio.NewReader(os.Stdin)
			result := ""
			for result != "y" && result != "n" {
				fmt.Println("Já existe uma biblioteca com a mesma URL e Hash. Deseja fazer executar a operação de PULL? [y/n]")
				r, err := reader.ReadString('\n')
				result = strings.ToLower(r)[:len(r)-1]
				check(err)
			}
			if result == "n" {
				nextStep = End
			} else {
				t := temp["libs"].(libs).Libs
				i := temp["repo"].(int)
				t = append(t[:i], t[i+1:]...)
				nextStep = PullRepo
			}
			break
		case PullRepo:
			repo, err := git.PlainOpen(filepath.Join(libsConfig.LibsDir, temp["Name"].(string)))
			check(err)
			wt, err := repo.Worktree()
			check(err)
			err = wt.Pull(&git.PullOptions{
				Progress: os.Stdout,
			})
			if err == git.NoErrAlreadyUpToDate {
				fmt.Println("Nenhuma atualização adicionada.")
				nextStep = End
				break
			}

			nextStep = CheckOutHash
			break
		case CheckOutHash:
			repo, err := git.PlainOpen(filepath.Join(libsConfig.LibsDir, temp["Name"].(string)))
			check(err)
			wt, err := repo.Worktree()
			check(err)

			err = wt.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(temp["Hash"].(string)),
			})
			check(err)

			nextStep = WriteToFile
			break

		case WriteToFile:
			llibs := temp["libs"].(libs)

			file, err := os.OpenFile(fmt.Sprintf("%s.%s", fnameLibs, "toml"), os.O_WRONLY|os.O_CREATE, 0644)
			check(err)

			llibs.Libs = append(llibs.Libs, lib{
				Hash:       temp["Hash"].(string),
				URL:        temp["URL"].(string),
				Name:       temp["Name"].(string),
				LastUpdate: time.Now(),
			})

			err = toml.NewEncoder(file).Encode(llibs)
			check(err)

			file.Close()

			fmt.Println("Libs escritas.")

			nextStep = End
			break

		case CreateDir:
			err := os.MkdirAll(filepath.Join(libsConfig.LibsDir, temp["Name"].(string)), os.ModePerm)
			check(err)

			nextStep = CloneToDir
			break

		case CloneToDir:
			_, err := git.PlainClone(filepath.Join(libsConfig.LibsDir, temp["Name"].(string)), false, &git.CloneOptions{
				URL:      temp["URL"].(string),
				Progress: os.Stdout,
			})
			check(err)

			nextStep = CheckOutHash
			break

		case End:
			os.Exit(0)
		}

	}

}

func main() {
	action := os.Args[1]
	switch action {
	case "init":
		initialize()
		break
	case "add":
		_, err := toml.DecodeFile(fmt.Sprintf("%s.%s", fnameCfg, "toml"), &libsConfig)
		check(err)
		addLib()
		break
	}
}
