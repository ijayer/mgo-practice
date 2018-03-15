/**
 * Created by @zheZh on 2016.12.9.
 */


db.production.aggregate(
    // Pipeline
    [
        // Stage 1
        {
            $project: {
                "_id": 1, "process_embed": 1
            }
        },

        // Stage 2
        {
            $unwind: "$process_embed"
        },

        // Stage 3
        {
            $match: {
                "process_embed._id": ObjectId("5840e6ac61016e2814fee5a2")
            }
        }

    ]

    // Created with 3T MongoChef, the GUI for MongoDB - http://3t.io/mongochef

);
