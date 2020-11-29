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

		destFile, err := os.Create(BaseUploadPath + header.Filename) // ?
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
		// filename exe_PartitionRes_prefix_maplerid
		filename := header.Filename
		Write2Shell("Start process file: " + filename)	
		mapleSource := MaplePath + filename
		// exe_prefix_subid
		exe := strings.Split(filename, "_")[0]
		maplerid := strings.Split(filename, "_")[3]
		Write2Shell("Mapler ID: " + maplerid)
		prefix := strings.Split(filename, "_")[2]
		exepath := ExePath + exe
		// TODO: map slow
		// step 1 process file with map and store it to a new file 
		intermediateFilename := "MapleIntermediate_" + prefix + "_" + maplerid
		cmd := exec.Command(exepath, mapleSource, intermediateFilename)
		_, err = cmd.Output()
		if err != nil {
			Logger.Fatal(err)
		}
		Write2Shell("Start split files")
		// step 2 split intermediate file based on keys
		// get all keys
		file, err := os.Open(intermediateFilename)
		if err != nil {
			Logger.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		allKeyList := make(map[string]string)
		for scanner.Scan() {
			text := scanner.Text()
			key := strings.Fields(text)[0] 
			if _, ok := allKeyList[key]; ok {
				continue
			} else {
				allKeyList[key] = ""
			}
		}
		if err := scanner.Err(); err != nil {
			Logger.Fatal(err)
		}
		file.Close()

		filepointerMap := make(map[string]*os.File)
		putfileList := []string{}
		for k, _ := range allKeyList {
			outputName := "mapleResult" + "_" + prefix + "_" + maplerid + "_" + k
			putfileList = append(putfileList, outputName)
			f, err := os.OpenFile(outputName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)	
			if err != nil {
				panic(err)
			}
			defer f.Close()
			filepointerMap[k] = f
		}

		file, err = os.Open(intermediateFilename)
		if err != nil {
			Logger.Fatal(err)
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			text := scanner.Text()
			key := strings.Fields(text)[0] 
			filepointerMap[key].WriteString(text + "\n")
		}
		if err := scanner.Err(); err != nil {
			Logger.Fatal(err)
		}

		for _, putfilename := range putfileList {
			client.PutFile(putfilename, putfilename)
			if err := os.Remove(putfilename); err != nil {
				Logger.Fatal(err)
			}
		}
		if err := os.Remove(mapleSource); err != nil {
			Logger.Fatal(err)
		}
		if err := os.Remove(intermediateFilename); err != nil {
			Logger.Fatal(err)
		}
		// at this time all maple results are on hdfs, no intermediate files are in datanodes
		Write2Shell("Mapler " + maplerid + " Success")
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
		Write2Shell("Juicer ID: " + id)

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
		workerIDkeyList := strings.Split(keys,"*")
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
		}

		// step 3. download all subfiles to \main \\ TODO:
		Write2Shell("Start download source files")
		for _, filename := range filenameList {
			// juice source small files mapleResult_prefix_maplerid_key
			//Write2Shell("try to download file: " + filename)
			client.GetFile(filename, filename)
		}

		//Write2Shell("finish download files")
		// merge those files to one key per file
		sourceFileList := []string{} // record all merged juice source files, after finish juice delete them
		for key, files := range key2fileMap {
			// use key to create a file
			juiceSourceFilename := "JuiceSource_" + prefix + "_" + key
			outFile, err := os.OpenFile(juiceSourceFilename, os.O_CREATE|os.O_WRONLY, 0644)
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
		Write2Shell("Start process files")
		destFilename := "juiceResult_" + prefix + "_" + id  // the destFile will be send to master /main folder, master will merge them
		destFile, err := os.OpenFile(destFilename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			Logger.Fatal(err)
		}
		destFile.Close()
		
		exepath := ExePath + exe
		for _, source := range sourceFileList {
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
			Write2Shell("KEY FINISH:" +  source)
		}

		Write2Shell("Start send result to master")
		// step 5. send the file to master, can we send it in the body? yes
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
		Write2Shell("Juicer " + id + " Suceess")
		return
	}
	http.HandleFunc("/juiceWorker", ProcessJuice)
}