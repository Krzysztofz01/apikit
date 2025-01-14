# Api-Kit

**The tool is currently in a state of very early development**

## About the project
The idea behind this tool is to solve a real-world problem that arose when testing networked and IoT devices that provide administration panels as HTTP services.

The tool, without the need to write any code, but by appropriately configuring the service with JSON configuration files, creates a proxy that makes HTTP requests to the specified subpages of the administration panels, providing the data gathered on them in JSON format. The project can therefore be described as a proxy wrapper, making static resources available via a REST API.