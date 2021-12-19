rs.initiate({ _id : "${id}",  members: [{ _id: 0, host: "${host}:27017" }]});
rs.status();
