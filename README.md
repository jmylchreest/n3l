# N3L - Nexus3 Latest [resolver]

This should be entirely unnecessary but at the moment there doesn't appear to be a convenient way to pull the latest of a given artifact in a given repo with nexus 3.0. It claims it'll come in 3.1, but until then this is a basic proxy service to allow tools such a puppet to do simple deployments using the latest or a specific version.

Usage is as follows:

* Execute the n3l binary on a host that is both accessible by clients and able to access the nexus3 repository
* On the client, query the n3l proxy with the following form:

`http http://n3lproxyserver:3001/fetch/{host}/{repo}/{group}/{artifact}/{version}/{extension}/{classifier}`

where:

* `{host}` is the hostname of your nexus 3.0 server
* `{repo}` is the repository name
* `{group}` is the maven group attribute
* `{artifact}` is the maven artifact name
* `{version}` is either a literal version string, or "latest"
* `{extension}` is the file suffix, such as `war`, `jar`, `pom`
* `{classifier}` is the maven artifact classifier
