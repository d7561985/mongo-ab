rs.initiate({ _id : "${id}",  members: [{ _id: 0, host: "${host}:50000" }]});
rs.status();
