package datanode

import (
	. "structs"
	"networking"
	"helper"
	"bufio"
	"os/exec"
	"client"
	"net/http"
	"log"
	"os"
	"io"
	"strings"
	"io/ioutil"
	"strconv"
)

func ServerRun(serverPort string) {
	// create a folder
	helper.CreateFile()

	// client puts; datanode downloads the file to local disk(hdfs folder)
	// endpoint /putfile
	networking.HTTPlistenDownload(Dir + "files_" + DatanodeHTTPServerPort + "/") 

	// master sends rereplica request to source; source sends file to datanode; datanode downloads file to local disk
	// endpoint /rereplica
	networking.HTTPlistenRereplica() 

	// new master sends recover request to datanode; datanode sends its local file list to master.
	// endpoint /recover
	networking.HTTPlistenRecover() 

	// client sends delete request to datanode; datanode deletes the file.
	// endpoint /deletefile
	networking.HTTPlistenDelete(Dir + "files_" + DatanodeHTTPServerPort + "/")

	// master sends maple source file to datanode; datanode store this file to local disk(hdfs folder) and run the program on this file.
	// endpoint /mapleWorker
	HTTPlistenMaple(Dir + "maplejuicefiles/") 

	// master sends juice source file to datanode; juicer will get the file to local disk /main 
	// endpoint /juiceWorker
	HTTPlistenJuice()

	// start above http services
	go networking.HTTPstart(DatanodeHTTPServerUploadPort) // start http server. main function: put, sub function: rereplica

	// client gets; datanode send the file to the client.
	go networking.HTTPfileServer(serverPort, Dir + "files_" + DatanodeHTTPServerPort) //handle get files
}

// worker (datanode) listen to maple request from master
func HTTPlistenMaple(BaseUploadPath string) {
	DownloadMaple := func(w http.ResponseWriter, r *http.Request) {
		// step 1. Download file
		formFile, header, err := r.FormFile("uploadfile")
		if err != nil {
			log.Printf("Get form file failed: %s\n", err)
			//TODO: w.write add return status
			w.Write([]byte("error"))
			return
		}
		defer formFile.Close()

		destFile, err := os.Create(BaseUploadPath + header.Filename)
		if err != nil {
			log.Printf("Create failed: %s\n", err)
			w.Write([]byte("error"))
			return
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, formFile)
		if err != nil {
			log.Printf("Write file failed: %s\n", err)
			w.Write([]byte("error"))
			return
		}
		
		// step 2. process the file
		// we assume the executable file is in the current folder
		filename := header.Filename
		mapleSource := MaplePath + filename
		// exe_prefix_subid
		exe := strings.Split(filename, "_")[0]
		exepath := ExePath + exe
		file, err := os.Open(mapleSource) 
		if err != nil {
			Logger.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		outputMap := make(map[string]string)
		for scanner.Scan() {
			text := scanner.Text()
			fields := strings.Fields(text)
			key := fields[0]
			value := fields[1]
			cmd := exec.Command(exepath, key, value)
			// TODO: here the newValue can be multi lines
			output, err := cmd.Output()
			if err != nil {
        		Logger.Fatal(err)
			}
			fields = strings.Fields(string(output))
			key = fields[0]
			outputMap[key] = outputMap[key] + string(output) 
		}
		if err := scanner.Err(); err != nil {
			Logger.Fatal(err)
		}
		for i,s := range outputMap {
			//name exe_prefix_subid_key
			outputName := "mapleResult" + "_" + filename + "_" + i
			err := ioutil.WriteFile(outputName, []byte(s), 0644)
			if err != nil {
				Logger.Fatal(err)
			}
			//TODO:we should add errror here 
			// here we change the remote name
			remotename := strings.Replace(outputName, "_"+exe, "", 1)
			client.PutFile(outputName, remotename)
			if err := os.Remove(outputName); err != nil {
				Logger.Fatal(err)
			}

		}
		if err := os.Remove(mapleSource); err != nil {
			Logger.Fatal(err)
		}
		w.Write([]byte("OK"))
	}
	http.HandleFunc("/mapleWorker", DownloadMaple)
}

//juice
func HTTPlistenJuice() {
	ProcessJuice := func(w http.ResponseWriter, req *http.Request) {
		// step 1. get all variables in the url
		prefixs, ok := req.URL.Query()["prefix"]
		if !ok {
			Logger.Error("Handle Juice Url Param 'prefix' is missing")
			return
		}
		prefix := prefixs[0] // exe_prefix
		exe := strings.Split(prefix, "_")[0]
		prefix = strings.Split(prefix, "_")[1]

		ids, ok := req.URL.Query()["id"]
		if !ok {
			Logger.Error("Handle Juice Url Param 'id' is missing")
			return
		}
		id:= ids[0] 
		Write2Shell(id)

		keyss, ok := req.URL.Query()["keys"]
		if !ok {
			Logger.Error("Handle Juice Url Param 'prefix' is missing")
			return
		}
		keys := keyss[0]
		// no key for this juicer. rare. may happen in hash shuffle
		if keys == "" {
			w.Write([]byte("OK"))
			return
		}

		// step 2. create a list for all to be downloaded files
		
		filenameList := []string{}
		key2fileMap := make(map[string][]string) // we use this data structure to merge files
		workerIDkeyList := strings.Split(keys,",")
		workerIDkeyList = workerIDkeyList[:len(workerIDkeyList)-1]
		for _, key := range workerIDkeyList{
			filename := "mapleResult_" + prefix + "_" + key
			filenameList = append(filenameList, filename)
			k := strings.Split(key, "_")[1]
			if ok:=key2fileMap[k]; ok == nil {
				key2fileMap[k] = []string{filename}
			} else {
				key2fileMap[k] = append(key2fileMap[k], filename)
			}
			Write2Shell(filename)
		}

		// step 3. download all subfiles to \main
		for _, filename := range filenameList {
			client.GetFile(filename, filename)
		}
		// merge those files to one key one file
		sourceFileList := []string{} // record all merged juice source files, after finish juice delete them
		for key, files := range key2fileMap {
			// use key to create a file
			juiceSourceFilename := "JuiceSource_" + prefix + "_" + key
			outFile, err := os.OpenFile(juiceSourceFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				Logger.Fatal(err)
			}
			defer outFile.Close()

			for _, file := range files {
				inFile, err := os.Open(file)
				if err != nil {
					Logger.Fatal(err)
				}
				_, err = io.Copy(outFile, inFile)
				if err != nil {
					Logger.Fatal(err)
				}
				inFile.Close()
				// remove those file after append
				if err := os.Remove(file); err != nil {
					Logger.Fatal(err)
				}
			}
			sourceFileList = append(sourceFileList, juiceSourceFilename)
		}

		// step 4. process those files and save it to a file juiceResult_prefix_juicerworkerid
		destFilename := "juiceResult_" + prefix + "_" + id  // the destFile will be send to master /main folder, master will merge them
		destFile, err := os.Create(destFilename)
		if err != nil {
			Logger.Fatal(err)
		}
		destFile.Close()
		
		exepath := ExePath + exe
		for _, source := range sourceFileList {
			Write2Shell(source)
			key := strings.Split(source, "_")[2]
			// source is in main
			cmd := exec.Command(exepath, key, source, destFilename)
			_, err := cmd.Output()
			if err != nil {
        		Logger.Fatal(err)
			}
			// remove those source files
			if err := os.Remove(source); err != nil {
				Logger.Fatal(err)
			}
		}

		// step 5. send the file to master, can we send it in the body?
		Openfile, err := os.Open(destFilename)
		if err != nil {
			Logger.Fatal(err)
		}
		destFile.Close()
		FileHeader := make([]byte, 512)
		//Copy the headers into the FileHeader buffer
		Openfile.Read(FileHeader)
		//Get content type of file
		FileContentType := http.DetectContentType(FileHeader)
		//Get the file size
		FileStat, _ := Openfile.Stat()                     //Get info from file
		FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string
		//Send the headers
		w.Header().Set("Content-Disposition", "attachment; filename="+destFilename)
		w.Header().Set("Content-Type", FileContentType)
		w.Header().Set("Content-Length", FileSize)

		//Send the file
		//We read 512 bytes from the file already, so we reset the offset back to 0
		Openfile.Seek(0, 0)
		io.Copy(w, Openfile) //'Copy' the file to the client

		if err := os.Remove(destFilename); err != nil {
			Logger.Fatal(err)
		}
		return


/**
		// datanode can decode exe 
		toSendPrefix := exe + "_" + prefix
		Write2Shell("prefix: " + prefix)
		// step 1. Download file
		formFile, header, err := r.FormFile("uploadfile")
		if err != nil {
			log.Printf("Get form file failed: %s\n", err)
			//TODO: w.write add return status
			w.Write([]byte("error"))
			return
		}
		defer formFile.Close()

		destFile, err := os.Create(BaseUploadPath + header.Filename)
		if err != nil {
			log.Printf("Create failed: %s\n", err)
			w.Write([]byte("error"))
			return
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, formFile)
		if err != nil {
			log.Printf("Write file failed: %s\n", err)
			w.Write([]byte("error"))
			return
		}
		
		// step 2. process the file
		// we assume the executable file is in the current folder
		filename := header.Filename
		mapleSource := MaplePath + filename
		// exe_prefix_subid
		exe := strings.Split(filename, "_")[0]
		exepath := ExePath + exe
		file, err := os.Open(mapleSource) 
		if err != nil {
			Logger.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		outputMap := make(map[string]string)
		for scanner.Scan() {
			text := scanner.Text()
			fields := strings.Fields(text)
			key := fields[0]
			value := fields[1]
			cmd := exec.Command(exepath, key, value)
			// TODO: here the newValue can be multi lines
			output, err := cmd.Output()
			if err != nil {
        		Logger.Fatal(err)
			}
			fields = strings.Fields(string(output))
			key = fields[0]
			outputMap[key] = outputMap[key] + string(output) 
		}
		if err := scanner.Err(); err != nil {
			Logger.Fatal(err)
		}
		for i,s := range outputMap {
			//name exe_prefix_subid_key
			outputName := "mapleResult" + "_" + filename + "_" + i
			err := ioutil.WriteFile(outputName, []byte(s), 0644)
			if err != nil {
				Logger.Fatal(err)
			}
			//TODO:we should add errror here 
			client.PutFile(outputName, outputName)
			if err := os.Remove(outputName); err != nil {
				Logger.Fatal(err)
			}

		}
		if err := os.Remove(mapleSource); err != nil {
			Logger.Fatal(err)
		}
		w.Write([]byte("OK"))
		**/
	}
	http.HandleFunc("/juiceWorker", ProcessJuice)
}