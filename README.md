# Terraform AzureRM Provider Analysis

A tool to analyze all Azure Go Sdk endpoints used in Terraform AzureRM provider, which serves the purpose to calculate resource coverage.

## How Does It Works

this tool is generally composed of two parts: extractor and server.

### extractor

extractor is responsible for extracting all Azure Go Sdk endpoints by scanning the source code. 

Benefit from the AST and type checker in Go's standard library, it's easy to check whether a node in AST is a Call Expression Node and whether the expression type is from Azure Go SDk. In this way, we could know which methods in Azure Go SDK are invoked.

Then by analyzing the Go SDK, filtered the invoked method. Because Azure GO SDK are generated from swagger file, there is a pattern for every method. So It's easy to get the endpoints info from every method.

### server

server is responsible for automatically update the endpoints info periodically and insert into MS Sql Server. 

We provide two ways to trigger it: the cron job event trigger and http request trigger.