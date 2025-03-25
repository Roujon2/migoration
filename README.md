# Migoration

## PostgreSQL Migration CLI Tool

### Overview
This project is a simple migration tool made in Golang for PostgreSQL databases. It is a command line tool that allows you to create, apply and rollback migrations.

### Installation
To install the CLI tool globally, follow these steps:

1. Clone the repository:
```bash
git clone https://github.com/Roujon2/migoration.git
cd migoration
```

2. You can build the project:
```bash
go build -o migoration
```
Or install it directly:
```bash
go install
```

3. To make the 'migoration' command available globally, move the binary to your Go bin path (if built manually):
```bash
mv migoration $GOPATH/bin
```
Make sure your Go bin path is in your PATH environment variable ($GOPATH/bin is the default path):
```bash
export PATH=$PATH:$HOME/go/bin
```

4. Verify the installation by running the following command:
```bash
migoration --help
```
