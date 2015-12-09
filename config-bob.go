package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/foomo/config-bob/builder"
	"github.com/foomo/config-bob/vault"
)

const helpCommands = `
Commands:
    build         my main task
    vault-local   set up a local vault
`

func help() {
	fmt.Println("usage:", os.Args[0], "<command>")
	fmt.Println(helpCommands)
}

const (
	commandBuild       = "build"
	commandVaultLocal  = "vault-local"
	commandVaultRemote = "vault-remote"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "bash":
			fmt.Println("enter your vault_token please")
			/*
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Split(bufio.ScanBytes)
				for scanner.Scan() {
					b := scanner.Bytes()
					if len(b) > 0 {
						log.Println(b)
					}
					//fmt.Println(scanner.Text()) // Println will add back the final '\n'
				}
				if err := scanner.Err(); err != nil {
					fmt.Fprintln(os.Stderr, "reading standard input:", err)
				}
			*/
			os.Setenv("FOO", "BAR")
			fmt.Println("setting env:", "foo", "bar")
			fmt.Println("starting a shell:", os.Getenv("SHELL"), "--login")
			cmd := exec.Command(os.Getenv("SHELL"), "--login")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			log.Println(err)
		case commandVaultRemote:
			// VAULT_ADDR
			// VAULT_TOKEN
		case commandVaultLocal:
			if len(os.Args) == 3 {
				vaultFolder := os.Args[2]
				vault.LocalSetEnv()
				if !vault.LocalIsSetUp(vaultFolder) {
					fmt.Println("setting up vault tree")
					err := vault.LocalSetup(vaultFolder) //StartAndInit(os.Args[2])
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}
				}
				if vault.LocalIsRunning() {
					fmt.Println("there is already a vault running aborting")
					os.Exit(1)
				}
				fmt.Println("vault not running - trying to start it")
				vaultCommand, chanVaultErr, vaultErr := vault.LocalStart(vaultFolder)
				if vaultErr != nil {
					fmt.Println("could not start local vault server:", vaultErr.Error())
					os.Exit(1)
				}

				log.Println("launching new shell", "\""+os.Getenv("SHELL")+"\"", "with pimped environment")

				cmd := exec.Command(os.Getenv("SHELL"), "--login")
				go func() {
					vaultRunErr := <-chanVaultErr
					cmd.Process.Kill()
					fmt.Println("vault died on us")
					if vaultRunErr != nil {
						fmt.Println("vault error", vaultRunErr.Error())
					}
				}()
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				runErr := cmd.Run()
				if runErr != nil {
					fmt.Println("shell exit:", runErr.Error())
				}
				killErr := vaultCommand.Process.Kill()
				if killErr != nil {
					log.Println("could not kill vault process:", killErr.Error())
				}
				fmt.Println("config bob says bye, bye")
			} else {
				fmt.Println("usage: ", os.Args[0], commandVaultLocal, "path/to/vault/folder")
				os.Exit(1)
			}
		case commandBuild:
			builderArgs, err := builder.GetBuilderArgs(os.Args[2:])
			if err != nil {
				fmt.Println()
				fmt.Println("build usage", err.Error())
				os.Exit(1)
			} else {
				result, err := builder.Build(builderArgs)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				writeError := builder.WriteProcessingResult(builderArgs.TargetFolder, result)
				fmt.Println(writeError)
			}
		default:
			help()
		}
	} else {
		help()
	}
}
