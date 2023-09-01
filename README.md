# CheckSys

The CheckSys is a simple tool to check the system status.

## Introduction

Developed in Go, the CheckSys is a simple tool to check the system status. It can be used to check the system status in a simple way.

Do you want to know the system status? Just run the CheckSys.

## Usage

### Download

Git clone the CheckSys to your local machine.

```bash
git clone <repo>
```

### Config file

Copy the .env.example to .env and edit it.

```bash
cp .env.example .env
```
set the config in .env file.

### Build

Aceess the CheckSys directory and build it.

```bash
cd <repo>
go build checksys
```

### Run

Run the CheckSys.

```bash
sudo ./checksys
```

Is necessary to run the CheckSys with sudo, because it needs to access some system files.

## License

[MIT License](LICENSE)



