# Family Tree

Learner project for playing with Neo4j

## Config
Add the following to neo4j.conf: 
- server.jvm.additional=-Djava.net.preferIPv4Stack=true
- dbms.security.procedures.unrestricted=bloom.*,apoc.*,gds.*
- dbms.security.procedures.allowlist=bloom.*,apoc.load.*,gds.*
- server.unmanaged_extension_classes=com.neo4j.bloom.server=/bloom
- dbms.security.http_auth_allowlist=/,/browser.*,/bloom.*
- dbms.bloom.authorization_role=admin,architect,reader,bloom
