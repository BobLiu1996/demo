# start redis
docker run -d --name redis -v /home/mystic/redis_data:/data -p 6379:6379 -e REDIS_PASSWORD=mystic redis
# start mysql
docker run -d --name mysql-container -e MYSQL_ROOT_PASSWORD=mystic -p 3306:3306 -v /home/mystic/mount_data:/var/lib/mysql mysql:latest