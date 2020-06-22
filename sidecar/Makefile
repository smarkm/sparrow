service=sparrow-sidecar-injector-service.sparrow-system.svc
run-controller:
	go run . -url https://172.16.2.101:6443 -config ~/.kube/config

gen-ca-key: #1
	openssl genrsa -out ca.key -des3 2048
gen-ca-crt: #2
	openssl req -new -x509 -key  ca.key -days 7300 -out ca.crt
gen-server-key:
	echo "create server key :server.key"
	openssl genrsa -des3 -out server.key 2048 && \
	openssl rsa -in server.key -out server-pri.key
gen-server-csr: gen-server-key
	echo "create server csr: server.csr"
	openssl req -new -key server-pri.key -subj "/CN=${service}/O=FP/C=AU" -out server.csr
gen-server-crt:
	echo "create server crt: server.crt"
	openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 825 -sha256 
	#-extfile dev.mergebot.com.ext
gen-ca-boudle:  gen-server-key gen-server-csr gen-server-crt
	echo "create ca boudle: server.ca"
	cat server.crt | openssl base64|tr -d '\n' >server.ca
gen-server-all:
	openssl req -subj "/CN=${service}/O=FP/C=AU" -x509 -nodes -days 365 -newkey  rsa:2048 -keyout server.key -out server.crt && \
	cat server.crt | openssl base64|tr -d '\n' >server.ca
rm-server-ca:
	rm -rf server*
deploy-test:
	kubectl create deployment nginx --image nginx:1.17  -n sparrow-system 	
	#-oyaml	