/**
 * Created by @zheZh on 2016.12.9.
 */

db.sewdata.aggregate(
    // Pipeline
    [
        // Stage 1
        {
            $match: {
                "date": {"$gte": "2016-11-30", "$lte": "2016-12-02"}, "mac": "AC:CF:23:B8:7A:43"
            }
        }

    ]

    // Created with 3T MongoChef, the GUI for MongoDB - http://3t.io/mongochef

);
