# **VM1: Web Server & Gateway**

This virtual machine (VM1) acts as the main entry point and user interface for the distributed system. It serves HTML pages to the user and acts as a client, coordinating with other internal microservices (VM2 and VM3) via **JSON-RPC** to fulfill user requests.

## **Implemented Features**

* **Centralized Gateway:** Provides the Login page and Dashboard. It does NOT perform authentication or store files locally.  
* **RPC Client:** Communicates seamlessly with Auth-VM for login verification and File-VM for retrieving secure media.  
* **Memory Monitoring (Pub/Sub Publisher):** Runs a background goroutine that periodically checks the server's RAM usage. If the allocated memory exceeds the predefined threshold (**300 MB**), it publishes an alert event to the Pub/Sub broker.  
* **Memory Leak Simulator:** Exposes a special endpoint (/consume-memory) designed to intentionally allocate memory and bypass the Garbage Collector. This is used strictly for testing the Pub/Sub alerting mechanism.

## **File Structure**

* main.go: The core web server, RPC dialing logic, and memory monitoring routines.  
* templates/: Contains the HTML interface (login.html, dashboard.html).  
* static/: Contains static assets like style.css.

## **Prerequisites & Configuration**

* Go installed on the operating system (Linux).  
* To connect to other VMs properly, you should pass their IP addresses via environment variables. If not provided, it defaults to 127.0.0.1.

Environment Variables:

* AUTH\_VM\_IP: IP and Port of VM2 (e.g., 192.168.1.10:8082)  
* FILE\_VM\_IP: IP and Port of VM3 (e.g., 192.168.1.11:8081)  
* PUBSUB\_IP: IP and Port of the Pub/Sub Subscriber/Broker (e.g., 192.168.1.12:8083)

## **How to Run**

1. Open a terminal and navigate to the service directory:  
   cd web-vm/

2. Export the IP addresses of the other VMs (Replace with your actual VM IPs):  
   export AUTH\_VM\_IP="\<VM2\_IP\>:8082"  
   export FILE\_VM\_IP="\<VM3\_IP\>:8081"  
   export PUBSUB\_IP="\<PUBSUB\_IP\>:8083"

3. Run the service using Go:  
   go run main.go

4. You will see the following output indicating the server has started:  
   VM1 (Web Server) is running on port 8080...

## **Testing the System**

* **Web Interface:** Open a browser and go to http://\<VM1\_IP\>:8080/.  
* **Trigger Memory Alert:** Once the subscriber is running, visit http://\<VM1\_IP\>:8080/consume-memory?mb=100 multiple times until the memory exceeds 300MB and the alert is triggered.