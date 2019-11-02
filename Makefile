vendor: vendor/create-dmg

vendor/create-dmg:
	rm -rf vendor/create-dmg
	git clone https://github.com/andreyvit/create-dmg vendor/create-dmg
	rm -rf vendor/create-dmg/.git

