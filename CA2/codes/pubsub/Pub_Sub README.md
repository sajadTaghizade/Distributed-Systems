# **Pub/Sub Monitoring Service**

This folder contains a lightweight Publish/Subscribe mechanism implemented over HTTP/JSON. It is designed to monitor the system's health, specifically to receive and display critical alerts when the Web Server (VM1) exceeds its memory allocation threshold.

## **Implemented Features**

* **HTTP-based Pub/Sub:** Acts as a simple event broker and subscriber, receiving JSON payloads and logging formatted alerts to the console.  
* **Standalone Publisher:** Includes a mock publisher.go script to independently verify the subscriber's functionality without needing to trigger a real memory leak on VM1.  
* **Clear Alerting:** The subscriber parses incoming MemoryEvent data and displays a highly visible, formatted warning in the terminal.

## **File Structure**

* subscriber.go: The main server that listens for incoming HTTP POST requests containing memory events and prints the alerts.  
* publisher.go: A mock script used strictly for testing. It sends a fake high-memory payload to the subscriber.

## **Prerequisites**

* Go installed on the operating system.  
* Ensure port 8083 is open on the machine running the subscriber to allow incoming connections from VM1.

## **How to Run the Subscriber**

1. Open a terminal and navigate to the directory:  
   cd pubsub/

2. Start the Subscriber service:  
   go run subscriber.go

3. You will see the following output, indicating it is waiting for events:  
   Pub/Sub Broker & Subscriber is listening on port 8083...

## **How to Test Independently (Mock Event)**

Before connecting VM1, you can test if the subscriber is receiving messages correctly:

1. Keep the subscriber.go running in one terminal.  
2. Open a **new** terminal in the same directory and run the mock publisher:  
   go run publisher.go

3. Check the subscriber's terminal. You should see a formatted URGENT ALERT displaying the mock memory usage details.

## **Integration with VM1**

During the full system test, VM1 will act as the publisher. When you hit the /consume-memory endpoint on VM1 and its RAM usage exceeds 300MB, it will automatically send an event to this subscriber. Make sure the PUBSUB\_IP environment variable in VM1 points to the IP and Port of the machine running this subscriber.