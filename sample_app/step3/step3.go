package main

import (
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

var users = make([]string, 100)

func generate_users(){
	for i := 0; i < len(users); i++ {
		users[i] = fmt.Sprintf("user%d",i)
	}
}

func get_followers(user string) []string {
	var id int
	fmt.Sscanf(user,"user%d",&id)
	followers := make([]string, 5)
	for i := 0; i < 5; i++ {
		followers[i] = users[(id+i+1)%len(users)]
	}
	return  followers
}


func insert_tweet(session *gocql.Session, user string, tweet_id gocql.UUID, tweet_time gocql.UUID, tweet_txt string){
	if err := session.Query(fmt.Sprintf("INSERT INTO tweets (user, tweet_id, time, text) VALUES ('%s',%s, %s,'%s')",
		user, tweet_id, tweet_time, tweet_txt)).Exec(); err != nil {
			log.Fatal(err)
	}


	for _, follower := range get_followers(user) {
		if err := session.Query(fmt.Sprintf("INSERT INTO timeline (user, tweet_id, time, author, text) VALUES ('%s',%s,%s,'%s','%s')",
			follower, tweet_id, tweet_time, user, tweet_txt)).Exec(); err != nil {
				log.Fatal(err)
		}
	}
}

func get_timeline(session *gocql.Session, user string) {
	var tweet_id gocql.UUID
	var author string
	var text string

	iter := session.Query(`SELECT tweet_id, author, text FROM timeline WHERE user = ?`, user).Iter()
	for iter.Scan(&tweet_id, &author, &text) {}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1", "127.0.0.2", "127.0.0.3")
	cluster.Keyspace = "example"
	cluster.Consistency = gocql.Quorum
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy());
	session, _ := cluster.CreateSession()
	defer session.Close()

	generate_users()

	// insert tweets
	for _, user := range users {
		for msg := 0 ; msg < 100; msg++ {
			insert_tweet(session, user, gocql.TimeUUID(), gocql.TimeUUID(), fmt.Sprintf("msg_%s_%d",user,msg))
		}
	}

	// fetch timelins
	for _, user := range users {
		get_timeline(session, user)
	}
}