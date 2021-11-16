let config = { _id: "rs1", members:[
        { _id : 0, host : "127.0.0.1:27021" ,priority : 100 },
        // { _id : 1, host : "127.0.0.1:27022"},
        // { _id : 2, host : "127.0.0.1:27023", arbiterOnly: true, priority: 0}
        ]
};
rs.initiate(config);
rs.status();
