This is the layout of the rvpay-go microservices server.

/protobuf
    This holds all the protos that will be used to generate the protobuf files needed for gRPC communication between services and clients. You should see three files:
        deposits.proto
        Dockerfile
        Makefile
    deposits.proto is going to generate the protobufs of the deposits service.
    Dockerfile is currently empty.
    Makefile contains the commands to lint and generate-protos from the .protos files.

/third-party
    This is a submodule to the googleapis directory. This does not concern you.

/grpc
    This is were all the generated protobuf files will be stored. It has a sub-folder /go
    /go
        This is where all golang related protobufs will be stored.
        /depositsgrpc
            This is where all the generated protobuf files for the deposits.proto are stored. This is done following the Makefile direction.

/deposits
    This is the deposits service folder.
    /cmd
        This is where the server and/or handlers, queues etc are run.
        /grpc
            This is the file that starts up the server and persists it. Understand the contents of this file well.

    /config
        This is where the shared structs are stored for the deposits service. Things like the Config struct which has a method LoadConfig to inject the environment variables into the Config object.

    /db
        This holds all db logic    
        /migrations
            All the migration files. Both up and down
        /query
            All the query files
        /repo
            Handles the db's peristence.
        /sqlc
            Holds the generated code for the queries.
        doc.go & sqlc.yaml aid in the conversion of the queries into sqlc.
    
    /deposits
        This holds all the service layer logic. This includes the service.go where all methods of the deposits service are stored.
        The service_test.go is currently empty.
    .env
        Holds all the environment variables required for the deposits service start up and persistence.
    Makefile
        Holds all the commands to run the deposits service.

Makefile:
    This is mostly to hold test commands. Don't mind it for now.

README.md
    README.md for the rvpay-go server.