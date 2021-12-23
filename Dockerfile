FROM mongo:5.0
ADD data/replica-init.js  /docker-entrypoint-initdb.d/replica-init.js
CMD [ "--logpath", "/dev/null", "--replSet", "rs",  "--bind_ip_all"]