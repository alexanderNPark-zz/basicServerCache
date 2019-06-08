package basicServerCache

import (
"flag"
"fmt"
"github.com/61c-teach/sp19-proj5-userlib"
"net/http"
"log"
"strings"
"time"


)

// This is the handler function which will handle every request other than cache specific requests.
func handler(w http.ResponseWriter, r *http.Request) {
	// FIXME This should be using the cache!
	// Note that we will be using userlib.ReadFile we provided to read files on the system.
	// The path to the file is given by r.URL.Path and will be the path to the string.
	// Make sure you properly sanitise it (more described in get file).
	/*** MODIFY THIS CODE ***/
	// Reads the file from the disk
	//filename := r.URL.Path[1:]
	response:=getFile(r.URL.Path)

	//PUT ERROR CHECK

	if(response.responseError!=nil){
		http.Error(w, userlib.FILEERRORMSG, userlib.FILEERRORCODE)
		return
	}
	if(response.responseData==nil){
		http.Error(w, userlib.TimeoutString, userlib.TIMEOUTERRORCODE)
		return
	}


	/*
	response, err := userlib.ReadFile(workingDir, filename)
	if err != nil {
		// If we have an error from the read, will return the generic file error message and set the error code to follow that.
		http.Error(w, userlib.FILEERRORMSG, userlib.FILEERRORCODE)
		return


    If you receive a timeout error, you should pass into the http.Error function the message userlib.TimeoutString with the return code as userlib.TIMEOUTERRORCODE.
    If the error is not a timeout, make sure you JUST pass into the http.Error function the string userlib.FILEERRORMSG with the return code userlib.FILEERRORCODE.

	}
*/

	// This will automatically set the right content type for the reply as well.
	w.Header().Set(userlib.ContextType, userlib.GetContentType(response.filename))
	// We need to set the correct header code for a success since we should only succeed at this point.
	w.WriteHeader(userlib.SUCCESSCODE) // Make sure you write the correct header code so that the tests do not fail!
	// Write the data which is given to us by the response.
	//w.Write(response)
	w.Write(response.responseData)
}

// This function will handle the requests to acquire the cache status.
// You should not need to edit this function.
func cacheHandler(w http.ResponseWriter, r *http.Request) {
	// Sets the header of the request to a plain text format since we are just dumping information about the cache.
	// Note that we are just putting a fake filename which will get the correct content type.
	w.Header().Set(userlib.ContextType, userlib.GetContentType("cacheStatus.txt"))
	// Set the success code to the proper success code since the action should not fail.
	w.WriteHeader(userlib.SUCCESSCODE)
	// Get the cache status string from the getCacheStatus function.
	w.Write([]byte(getCacheStatus()))
}

// This function will handle the requests to clear/restart the cache.
// You should not need to edit this function.
func cacheClearHandler(w http.ResponseWriter, r *http.Request) {
	// Sets the header of the request to a plain text format since we are just dumping information about the cache.
	// Note that we are just putting a fake filename which will get the correct content type.
	w.Header().Set(userlib.ContextType, userlib.GetContentType("cacheClear.txt"))
	// Set the success code to the proper success code since the action should not fail.
	w.WriteHeader(userlib.SUCCESSCODE)
	// Get the cache status string from the getCacheStatus function.
	w.Write([]byte(CacheClear()))
}

// The structure used for responding to file requests.
// It contains the file contents (if there is any)
// or the error returned when accessing the file.
// Note that it is only used by you so you do not
// need to use all of the fields in it.
type fileResponse struct {
	filename string
	responseData []byte
	responseError error
	responseChan chan *fileResponse
}

// To request files from the cache, we send a message that
// requests the file and provides a channel for the return
// information.
// Note that it is only used by you so you do not
// need to use all of the fields in it.
type fileRequest struct {
	filename string
	response chan *fileResponse
}

// DO NOT CHANGE THESE NAMES OR YOU WILL NOT PASS THE TESTS
// Port of the server to run on
var port int
// Capacity of the cache in Bytes
var capacity int
// Timeout for file reads in Seconds.
var timeout int
// The is the working directory of the server
var workingDir string

// The channel to pass file read requests to. This is how you will get a file from the cache.
var fileChan = make(chan *fileRequest)
// The channel to pass a request to get back the capacity info of the cache.
var cacheCapacityChan = make(chan chan string)
// The channel where a bool passed into it will cause the OperateCache function to be closed and all of the data to be cleared.
var cacheCloseChan = make(chan bool)

// A wrapper function that does the actual getting of the file from the cache.
func getFile(filename string) (response *fileResponse) {
	// You need to add sanity checking here: The requested file
	// should be made relative (strip out leading "/" characters,
	// then have a "./" put on the start, and if there is ever the
	// string "/../", replace it with "/", the string "\/" should
	// be replaced with "/", and finally any instances of "//" (or
	// more) should be replaced by a single "/".
	// Hint: A replacement may lead to needing to do more replacements!

	// Also if you get a request which is just "/", you should return the file "./index.html"

	// You should also return a timeout error (take a look at the userlib) after `timeout`
	// seconds if there is no response from the disk.

	/*** YOUR CODE HERE ***/
	temp:=filename
	a:=true
	for a{
		a=false
		filename = strings.Replace(filename,"/../", "//", -1)
		filename = strings.Replace(filename,"\\/", "//", -1)
		filename = strings.Replace(filename,"//", "/", -1)
		if(filename!=temp){
			a = true
			temp = filename
		}
	}
	if(strings.HasSuffix(filename,"/")){
		filename+="index.html"
	}
	if(strings.HasPrefix(filename,"/")){
		filename = "."+filename
	}else{
		filename = "./"+filename
	}

	/*** YOUR CODE HERE END ***/

	// The part below will make a request on the fileChan and wait for a response to be issued from the cache.
	// You should not really need to modify anything below here.
	// Makes the file request object.
	request := fileRequest{filename, make(chan *fileResponse)}
	// Sends a pointer to the file request object to the fileChan so the cache can process the file request.
	fileChan <- &request
	// Returns the result (from the fileResponse channel)
	return <- request.response
}

// This function returns a string of the cache current status.
// It will just make a request to the cache asking for the status.
// You should not need to modify this function.
func getCacheStatus() (response string) {
	// Make a channel for the response of the Capacity request.
	responseChan := make(chan string)
	// Send the response channel to the capacity request channel.
	cacheCapacityChan <- responseChan
	// Return the reply.
	return <- responseChan
}

// This function will tell the cache that it needs to close itself.
// You should not need to modify this function.
func CacheClear() (response string) {
	// Send the response channel to the capacity request channel.
	cacheCloseChan <- true
	// We should only return to here once we are sure the currently open cache will not process any more requests.
	// This is because the close channel is blocking until it pulls the item out of there.
	// Now that the cache should be closed, lets relaunch the cache.
	go operateCache()
	return userlib.CacheCloseMessage
}


type cacheEntry struct {
	filename string
	data []byte

	// You may want to add other stuff here...
	length int
	//beingWritten bool
	//finishedWriting chan bool


}

var entryUse = make(map[string]chan bool) // experiment last hope

// This function is where you need to do all the work...
// Basically, you need to...

// 1:  Create a map to store all the cache entries.

// 2:  Go into a continual select loop.

// Hint, you are going to want another channel

func operateCache() {
	/* TODO Initialize your cache and the service requests until the program exits or you receive a message on the
	 * cacheCloseChan at which point you should clean up (aka clear any caching global variables and return from
	 * this function. */
	// HINT: Take a look at the global channels given above!
	/*** YOUR CODE HERE ***/
	// Make a file map (this is just like a hashmap in java) for the cache entries.
	fileMap:=make(map[string]cacheEntry)
	numOfEntries:=0
	totalBytes:=0
	var c = make(chan int, 1)
	var asyncFile = make(chan int, 1)
	c<-1
	asyncFile<-1

	handleReq:=func(fileReqU *fileRequest){
		found:=make(chan *fileResponse)

		parallelRet:=func() {
			entry, inCache1:=fileMap[fileReqU.filename]
			if(inCache1){

				found<-&fileResponse{fileReqU.filename, entry.data, nil,nil}
				return
			}
			content, err := userlib.ReadFile(workingDir, fileReqU.filename)
			length:=len(content)
			if(err!=nil){
				found<-&fileResponse{fileReqU.filename, content, err,nil}
				return
			}
			<-c
			entry2, inCache2:=fileMap[fileReqU.filename]
			if(inCache2){
				found<-&fileResponse{fileReqU.filename, entry2.data, nil,nil}
				c<-1
				return
			}
			if(length>capacity){
			} else if(totalBytes+length>capacity){
				for key,val:=range fileMap{
					if(totalBytes+length<=capacity){
						break;
					}
					totalBytes-=val.length
					delete(fileMap, key)
					numOfEntries--;
				}

				numOfEntries+=1
				totalBytes+=length
				fileMap[fileReqU.filename] = cacheEntry{fileReqU.filename, content,length}
			}else{
				numOfEntries+=1
				totalBytes+=length
				fileMap[fileReqU.filename] = cacheEntry{fileReqU.filename, content,length}
			}
			c<-1
			found<-&fileResponse{fileReqU.filename, content, err,nil}
		}
		go parallelRet()

		select {
		case content:=<-found:
			fileReqU.response<-content
		case <- time.After(time.Duration(timeout)*time.Second):

			fileReqU.response<-&fileResponse{fileReqU.filename,nil, nil ,nil}
		}
	}
	// Once you have made a filemap, here is a good skeleton for you to use to handle requests.
	for {
		// We want to select what we want to do based on what is in different cache channels.
		select {
		case fileReq := <- fileChan:
			// Handle a file request here.
			entryFound, inCache:=fileMap[fileReq.filename]
			if(inCache){
				fileReq.response<-&fileResponse{fileReq.filename, entryFound.data, nil,nil}
			} else{
				go handleReq(fileReq) //disk read
			}
		case cacheReq := <- cacheCapacityChan:
			cacheReq<- fmt.Sprintf(userlib.CapacityString, numOfEntries, totalBytes, capacity)
		case <- cacheCloseChan:
			return
		}
	}
	/*** YOUR CODE HERE END ***/
}


// This functions when you do `go run server.go`. It will read and parse the command line arguments, set the values
// of some global variables, print out the server settings, tell the `http` library which functions to call when there
// is a request made to certain paths, launch the cache, and finally listen for connections and serve the requests
// which a connection may make. When it services a request, it will call one of the handler functions depending on if
// the prefix of the path matches the pattern which was set by the HandleFunc.
// You should not need to modify any of this.
func main(){
	// Initialize the arguments when the main function is ran. This is to setup the settings needed by
	// other parts of the file server.
	flag.IntVar(&port, "p", 8080, "Port to listen for HTTP requests (default port 8080).")
	flag.IntVar(&capacity, "c", 100000, "Number of bytes to allow in the cache.")
	flag.IntVar(&timeout, "t", 2, "Default timeout (in seconds) to wait before returning an error.")
	flag.StringVar(&workingDir, "d", "public_html/", "The directory which the files are hosted in.")
	// Parse the args.
	flag.Parse()
	// Say that we are starting the server.
	fmt.Printf("Server starting, port: %v, cache size: %v, timout: %v, working dir: '%s'\n", port, capacity, timeout, workingDir)
	serverString := fmt.Sprintf(":%v", port)

	// Set up the service handles for certain pattern requests in the url.
	http.HandleFunc("/", handler)
	http.HandleFunc("/cache/", cacheHandler)
	http.HandleFunc("/cache/clear/", cacheClearHandler)

	// Start up the cache logic...
	go operateCache()

	// This starts the web server and will cause it to continue to listen and respond to web requests.
	log.Fatal(http.ListenAndServe(serverString, nil))
}
