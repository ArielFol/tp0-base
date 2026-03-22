import sys
import yaml

def main():
    output_file = sys.argv[1]
    num_clients = int(sys.argv[2])

    compose = {
        "name": "tp0",
        "services": {
            "server": {
                "container_name": "server",
                "image": "server:latest",
                "entrypoint": "python3 /main.py",
                "environment": [
                    "PYTHONUNBUFFERED=1",
                    "LOGGING_LEVEL=DEBUG"
                ],
                "networks": ["testing_net"],
                "volumes": [
                    "./server/config.ini:/config.ini"
                ]
            }
        },
        "networks": {
            "testing_net": {
                "ipam": {
                    "driver": "default",
                    "config": [
                        {"subnet": "172.25.125.0/24"}
                    ]
                }
            }
        }
    }

    for i in range(1, num_clients + 1):
        client_name = f"client{i}"
        compose["services"][client_name] = {
            "container_name": client_name,
            "image": "client:latest",
            "entrypoint": "/client",
            "environment": [
                f"CLI_ID={i}",
                "CLI_LOG_LEVEL=DEBUG"
            ],
            "networks": ["testing_net"],
            "volumes": [
                    "./client/config.ini:/config.yaml"
                ],
            "depends_on": ["server"]
        }
    with open(output_file, 'w') as f:
        yaml.dump(compose, f, sort_keys=False)
    
    print(f"Archivo '{output_file}' generado con {num_clients} clientes.")

if __name__ == "__main__":
    main()