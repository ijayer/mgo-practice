/**
 * Created by @zheZh on 2016.12.9.
 */

db.plan.aggregate(
    // Pipeline
    [
        // Stage 1
        {
            $match: {
                "_id": ObjectId("584533a47d89971ad460daa1"), "actived_status": true
            }
        },

        // Stage 2
        {
            $project: {
                "line": 1, "aim_count": 1, "line_num": 1, "process_num": 1
            }
        },

        // Stage 3
        {
            $unwind: "$line"
        },

        // Stage 4
        {
            $match: {
                "line.id": "584533d07d89971ad460daa2"
            }
        },

        // Stage 5
        {
            $unwind: "$line.process"
        },

        // Stage 6
        {
            $match: {
                "line.process.id": "584533d07d89971ad460daa4"
            }
        },

        // Stage 7
        {
            $unwind: "$line.process.machine"
        },

        // Stage 8
        {
            $match: {
                "line.process.machine.name": "machine_one"
            }
        }

    ]

    // Created with 3T MongoChef, the GUI for MongoDB - http://3t.io/mongochef

);
