# **VM3: File & Image Server**

This virtual machine (VM3) acts as a centralized repository for files and images. Instead of using a basic HTTP server, it provides a highly efficient **JSON-RPC** interface for the web service (VM1) to fetch media securely over the network.

## **Implemented Features**

* **RPC-Based File Transfer (Design Bonus):** Fulfilling the extra credit requirement, this service transfers raw file bytes (\[\]byte) directly through JSON-RPC rather than relying on standard HTTP static file serving.  
* **Path Traversal Protection:** The server strictly sanitizes incoming file requests using filepath.Base(). This security measure prevents malicious actors from using ../ payloads to access unauthorized directories outside the designated folders.  
* **Categorized Storage:** Requests are dynamically routed to either the ./images/ or ./files/ directory based on the requested category, keeping the server's storage organized.

## **File Structure**

* main.go: The JSON-RPC server and file-reading logic.  
* images/: Directory containing sample image files (e.g., sample.png, logo.jpg).  
* files/: Directory containing sample text or document files (e.g., sample.txt).

## **Prerequisites**

* Go installed on the operating system (Linux).  
* Ensure port 8081 is open on VM3's firewall to allow incoming TCP connections from VM1.  
* You must create the images and files directories and place at least one dummy file in them before running the server.

## **How to Run**

1. Open a terminal and navigate to the service directory:  
   cd file-vm/

2. Run the service using Go:  
   go run main.go

3. Upon successful execution, you will see:  
   VM3 (File RPC Server) is running on port 8081...

## **Networking Notes for Testing**

To display files on the web interface, VM1 must be configured with the IP address of VM3. VM1 will send a JSON-RPC request containing the Category and Filename to VM3 on port 8081, and VM3 will respond with the raw file data.