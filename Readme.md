# how to run??

# please install first
proto-tools -cmd=install
# generate proto from your database
proto-tools -cmd=gen-proto-db host=localhost name=transaction user=root password=

# PR
- Schema generator
    - Read Comment
        Parsing get required hit
    - Table Comment
        Whitelist parsing
- lewati DO_NOT_REMOVE when re generate
- add options whitelist
- joins Table
- get foreign key
- enchance svc
- add whitelist
- add option db name- 

git clone git@github.com:golang/protobuf.git && (cd protobuf && git checkout v1.2.0 && go build -o $GOBIN/protoc-gen-go ./protoc-gen-go) && rm -r protobuf

# how to deploy kube
-cmd=deploy mode=deployment svc=transaction env=production version=v1

-cmd=deploy mode=service svc=transaction env=production version=v1

-cmd=deploy mode=gateway svc=transaction env=production version=v1

-cmd=deploy mode=virtual svc=transaction env=production version=v1

-cmd=deploy mode=scale svc=transaction env=production version=v1

export GOOS=linux
go build -o jamet

PRAKERJA_ENV = production

./jamet -cmd=deploy mode=publish_digest svc=bank-integration env=production version=v1 digest=f82dbdd0
./jamet -cmd=deploy mode=push_env svc=bank-integration env=production version=v1 name=PRAKERJA_ENV value=production
./jamet -cmd=deploy mode=init svc=bank-integration env=production version=v1
./jamet -cmd=deploy mode=deployment svc=bank-integration env=production version=v1
./jamet -cmd=deploy mode=scale svc=bank-integration env=production version=v1 min=5 max=20

./jamet -cmd=deploy mode=deployment svc=partner-api env=production version=v1

./jamet -cmd=deploy mode=publish_digest svc=bank-integration env=production version=v1 digest=3624647

./jamet -cmd=deploy mode=deployment svc=transaction env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=bank-integration env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=batch env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=region env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=platform-api env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=wallet env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=user-api env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=survey-incentive env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=notification env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=image env=production version=v1 && \

	
./jamet -cmd=deploy mode=deployment svc=transaction env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=bank-integration env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=batch env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=region env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=platform-api env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=wallet env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=user-api env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=survey-incentive env=production version=v1 && \
./jamet -cmd=deploy mode=deployment svc=notification env=production version=v1 
	
	
	
	
wallet	


kubectl get pod -l app=bank-integration
kubectl logs -l app=bank-integration -c bank-integration

172.28.23.160

registry-intl.ap-southeast-5.aliyuncs.com/prakerja/wallet:69b4cc8

172.28.14.8

./jamet -cmd=deploy mode=publish_svc svc=survey-incentive env=production version=v1 min=3 max=10 host=survey-incentive.prakerja.local && \
./jamet -cmd=deploy mode=publish_digest svc=survey-incentive env=production version=v1 digest=f82dbdd0
./jamet -cmd=deploy mode=push_env svc=survey-incentive env=production version=v1 name=DBADDRESS value=pgm-d9jh549c9574ds9669900.pgsql.ap-southeast-5.rds.aliyuncs.com && \
./jamet -cmd=deploy mode=push_env svc=survey-incentive env=production version=v1 name=DBUSER value=pmo_survey && \
./jamet -cmd=deploy mode=push_env svc=survey-incentive env=production version=v1 name=DBNAME value=survey_production && \
./jamet -cmd=deploy mode=push_env svc=survey-incentive env=production version=v1 name=DBPASSWORD value=Pr4K3rj4S3laM4ny4 && \
./jamet -cmd=deploy mode=push_env svc=survey-incentive env=production version=v1 name=DBPORT value=1921 && \
./jamet -cmd=deploy mode=push_env svc=survey-incentive env=production version=v1 name=USER_API_ADDRESS value=user-api.default.svc.cluster.local && \
./jamet -cmd=deploy mode=push_env svc=survey-incentive env=production version=v1 name=TRANS_ADDRESS value=http://transaction.default.svc.cluster.local/api/v1/tr/survey/create

# clear docker logs 
echo "" > $(docker inspect --format='{{.LogPath}}' engine)

kubectl autoscale deployment transaction --min=10 --max=40
kubectl scale --current-replicas=2 --replicas=3 deployment/mysql  # jika Deployment bernama mysql saat ini memiliki ukuran 2, skalakan mysql menjadi 3

kubectl exec -it shell-demo -- /bin/bash
kubectl rollout restart deployment mydeploy

https://www.digitalocean.com/community/tutorials/how-to-do-canary-deployments-with-istio-and-kubernetes
https://www.digitalocean.com/community/tutorials/how-to-install-and-use-istio-with-kubernetes#step-2-%E2%80%94-installing-istio-with-helm

./jamet -cmd=deploy mode=init svc=survey-incentive env=production version=v1

go run main.go -cmd=etl host=rm-d9j026wg1l28z60gp66640.mysql.ap-southeast-5.rds.aliyuncs.com user=db_admin password=Pr4K3rj4S3laM4ny4 db=transaction kafka=172.28.14.199:9092 server=dbserver_trx group=trx_transaction_group

docker run  -dit --restart unless-stopped -e 'ROLE=data_mesh' -e 'CODE=certificate' -e 'MAX_ROUTINE=20' --name certificate  -d zokypesch/data_mesh:latest

go run main.go -cmd=etl host=rm-d9j026wg1l28z60gp66640.mysql.ap-southeast-5.rds.aliyuncs.com user=db_admin password=Pr4K3rj4S3laM4ny4 db=users kafka=172.28.14.199:9092 server=dbserver_users group=users_group

go run main.go -cmd=etl host=rm-d9js218880l8oh31a66640.mysql.ap-southeast-5.rds.aliyuncs.com user=batch_svc password=Pr4K3rj4S3laM4ny4 db=batch kafka=172.28.14.201:9092 server=dbserver_batch group=batch_group

go run main.go -cmd=etl host=rm-d9j99s2po8fycw3t366640.mysql.ap-southeast-5.rds.aliyuncs.com user=pmo_user password=Pr4K3rj4S3laM4ny4 db=pmo-platform kafka=172.28.14.200:9092 server=dbserver_general group=general_group

go run main.go -cmd=etl host=rm-d9j99s2po8fycw3t366640.mysql.ap-southeast-5.rds.aliyuncs.com user=pmo_user password=Pr4K3rj4S3laM4ny4 db=wallet_prakerja kafka=172.28.14.200:9092 server=dbserver_general group=general_group

go run main.go -cmd=etl host=rm-d9j99s2po8fycw3t366640.mysql.ap-southeast-5.rds.aliyuncs.com user=pmo_user password=Pr4K3rj4S3laM4ny4 db=survey_tests kafka=172.28.14.200:9092 server=dbserver_general group=general_group