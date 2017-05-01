package main

import (
	"fmt"
	"github.com/modmuss50/FactorioWrapper/utils"
	"github.com/modmuss50/FactorioWrapper/config"
	"os"
	"bufio"
	"io"
	"log"
	"strings"
	"time"
	"os/exec"
)

var (
	input io.WriteCloser
)



func main() {
	fmt.Println("Starting wrapper")

	dataDir := "./data/"
	if !utils.FileExists(dataDir){
		utils.MakeDir(dataDir)
	}

	fmt.Println("Loading config")
	config.LoadConfig()

	version := config.FactorioVersion
	processDir := utils.GetProcessDir(version)

	utils.HandleDownload(dataDir, version)

	fmt.Println("Starting game...")
	factorioProcess := getExec(processDir)
	fmt.Println("Getting input for game")
	factorioInput, err := factorioProcess.StdinPipe()
	input = factorioInput
	utils.TextInput = factorioInput
	if err != nil {
		log.Fatal(err)
	}
	defer factorioInput.Close()
	factorioOutput, _ := factorioProcess.StdoutPipe()


	scanner := bufio.NewScanner(factorioOutput)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			fmt.Printf("\t > %s\n", text)
			if strings.Contains(text, "changing state from(CreatingGame) to(InGame)") {
				utils.LoadDiscord(config.DiscordToken)
				utils.ChannelID = config.DiscordChannel
				utils.SendStringToDiscord("Server started on factorio version " + version, config.DiscordChannel)
			}
			if strings.Contains(text, "changing state from(CreatingGame) to(InitializationFailed)") || strings.Contains(text, "Couldn't acquire exclusive lock for") {
				fmt.Println("Game failed to start")
				factorioProcess.Process.Kill()
				os.Exit(1)
			}
			if strings.Contains(text, "[JOIN]") {
				utils.SendStringToDiscord(text[26:], config.DiscordChannel)
			}
			if strings.Contains(text, "[CHAT]") {
				if !strings.Contains(text, " [CHAT] <server>:") {
					utils.SendStringToDiscord(text[26:], config.DiscordChannel)
				}
			}
			if strings.Contains(text, "[LEAVE]") {
				utils.SendStringToDiscord(text[27:], config.DiscordChannel)
			}
			if strings.Contains(text, "Goodbye") {
				utils.SendStringToDiscord("Server closed", config.DiscordChannel)
				time.Sleep(1 * time.Second)
				os.Exit(0)
			}
		}
	}()

	fmt.Println("Launching process")
	er := factorioProcess.Start()
	if er != nil {
		log.Fatal(er)
		os.Exit(1)
	}

	//ticker := time.NewTicker(time.Second * 10)
	//go func() {
	//	for range ticker.C {
	//		//io.WriteString(factorioInput, "hello is this working?\n")
	//	}
	//}()

	readInput(factorioProcess)

}

func readInput(cmd *exec.Cmd) {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if strings.HasPrefix(text, "stop") {
		fmt.Println("Stopping server...")
		utils.KillProcess(cmd)

	}
	if strings.HasPrefix(text, "fstop") {
		fmt.Println("Stopping server")
		cmd.Process.Kill()
		return
	}

	ranCmd := false
	if strings.HasPrefix(text, "cmd") {
		ranCmd = true
		io.WriteString(input, strings.Replace(text, "cmd ", "", -1))
	}
	if !ranCmd {
		fmt.Println("Command not found!")
	}

	readInput(cmd)
}




func getExec(dir string) *exec.Cmd {
	fullDir := utils.GetBinPath()
	fullDir = utils.FormatPath(fullDir)
	fmt.Println(fullDir)
	factorioExec := exec.Command(fullDir, "--start-server", "./saves/" + config.FactorioSaveFileName)
	factorioExec.Dir = "." + dir
	return factorioExec
}
