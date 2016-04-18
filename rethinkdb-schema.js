r.dbCreate("been_there");

r.db("been_there").tableCreate("user_visits");
r.db("been_there").table("user_visits").indexCreate("user");

r.db("been_there").tableCreate("cities");
r.db("been_there").table("cities").indexCreate("state");
