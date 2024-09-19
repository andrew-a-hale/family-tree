default:
	go run -C generator main.go ;
	sudo neo4j-admin database import full --nodes=imports/people.csv --nodes=imports/generations.csv --relationships=imports/edges.csv --overwrite-destination --verbose
	rm -f import.report
	sudo neo4j console
