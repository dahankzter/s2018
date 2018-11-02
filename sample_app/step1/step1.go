package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

var (
	hosts = flag.String("hosts", "", "comma separated list of hosts to connect to")
	users = make([]string, 1000)
)

func generate_users() {
	for i := 0; i < len(users); i++ {
		users[i] = fmt.Sprintf("user%d", i)
	}
}

func get_followers(user string) []string {
	var id int
	fmt.Sscanf(user, "user%d", &id)
	followers := make([]string, 5)
	for i := 0; i < 5; i++ {
		followers[i] = users[(id+i+1)%len(users)]
	}
	return followers
}

func insert_tweet(session *gocql.Session, user string, tweet_id gocql.UUID, tweet_time gocql.UUID, tweet_txt string) {
	if err := session.Query(fmt.Sprintf("INSERT INTO tweets (user, tweet_id, time, text) VALUES ('%s',%s, %s,'%s')",
		user, tweet_id, tweet_time, tweet_txt)).Exec(); err != nil {
		log.Fatal(err)
	}

	for _, follower := range get_followers(user) {
		liked := false
		if rand.Intn(100) < 5 {
			liked = true
		}
		if err := session.Query(fmt.Sprintf("INSERT INTO timeline (user, tweet_id, time, author, text, liked) VALUES ('%s',%s, %s,'%s','%s', %t)",
			follower, tweet_id, tweet_time, user, tweet_txt, liked)).Exec(); err != nil {
			log.Fatal(err)
		}
	}
}

func get_timeline(session *gocql.Session, user string) {
	var tweet_id gocql.UUID
	var author string
	var text string

	iter := session.Query(fmt.Sprintf("SELECT tweet_id, author, text FROM timeline WHERE user = '%s'", user)).Iter()
	for iter.Scan(&tweet_id, &author, &text) {
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {

	flag.Parse()

	if *hosts == "" {
		flag.Usage()
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster(strings.Split(*hosts, ",")...)
	cluster.Keyspace = "scylla_demo"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()
	defer session.Close()

	generate_users()

	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	rate := time.Second / 20
	throttle := time.Tick(rate)
	for true {
		<-throttle
		user := users[random.Intn(len(users))]
		if random.Intn(10) > 5 {
			for msg := 0; msg < 100; msg++ {
				insert_tweet(session, user, gocql.TimeUUID(), gocql.TimeUUID(), fmt.Sprintf("msg_%s_%d", user, msg))
			}
		} else {
			get_timeline(session, user)
		}
	}
}
