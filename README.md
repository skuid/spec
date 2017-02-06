# Prefect

An ops document management system


**prefect** |ˈprēˌfekt|

noun

1. a chief officer, magistrate, or regional governor in certain countries: the prefect of police.
    * a senior magistrate or governor in the ancient Roman world: _Avitus was prefect of Gaul from AD 439._
2. chiefly Brit. in some schools, a senior student authorized to enforce discipline.


## Go must be installed

```
brew install go
mkdir -p ~/go/{src,bin,pkg}
export GOPATH=~/go
# Append GOPATH to profile
echo 'export GOPATH=~/go' | tee -a ~/.profile
```

## Installation

```bash
mkdir -p $GOPATH/src/bitbucket.org/productengineering/prefect
git clone git@bitbucket.org:productengineering/prefect.git $GOPATH/src/bitbucket.org/productengineering/prefect
cd $GOPATH/src/bitbucket.org/productengineering/prefect

# Install dependencies
go get ./...
```

## Usage

By default, the server listens on port 3000.

```
docker-compose up -d

# Set parameters for prefect to connect to the postgres database
export PGDATABASE="prefect"
export PGHOST="192.168.99.100"
export PGPASSWORD="prefect"
export PGUSER="prefect"

# See the help menu for configurable options
go run main.go -h
# or to make a binary
go build

```

### Seeds

Before you run the server, you'll need to seed the database. For ease,
a small cleanup script is present in `/hack/wipe.sh`

```
./prefect seed
```

### Adding users

```
./prefect adduser <username> <password> --email <email>
```

### Running the server
```
./prefect serve --ssl
```

Then open [https://localhost:3000/](https://localhost:3000/) to get to the site!

## Usage with Pliny

You'll need to use the `--ssl` flag on the server, and make sure your data source
is _not_ in proxy mode.

## Workflow

1. A document is created
    1. A DocumentTemplateVersion is created to store that document
1. 
