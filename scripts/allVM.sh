#! /bin/bash
# up address except vm 
IPAddress=(
        "172.22.158.12"
        "172.22.94.12"
        "172.22.156.13"
        "172.22.158.13"
        "172.22.94.13"
        "172.22.156.14"
        "172.22.158.14"
        "172.22.94.14"
        "172.22.156.15"
)
VM1IPAddress="172.22.156.12"
copyPublicKey="copyKey"

if [ "$1" == "alias" ]; then
	count=2
	for i in ${!IPAddress[@]};	
	do
			
		echo "alias vm$count='ssh -p 22 beitong2@${IPAddress[$i]}'" >> ~/.bashrc
		count=$((count+1))
	done
fi 

if [ "$1" == "$copyPublicKey" ]; then
	for i in ${!IPAddress[@]};	
	do
			
		host='beitong2@'${IPAddress[$i]}
		ssh-copy-id $host
	done
	host='beitong2@'$VM1IPAddress
	ssh-copy-id $host
else
	for i in ${!IPAddress[@]};	
	do
			
		host='beitong2@'${IPAddress[$i]}
		echo $host
		ssh $host 'bash -s' < $1
		echo "done"
	done
	if [ $1 == "runMp0.sh" ] || [ $1 == "setUpMp0.sh" ]
	then
		host='beitong2@'$VM1IPAddress
		echo $host
		ssh $host 'bash -s' < $1
		echo "done"
	fi
		
		 
fi 
