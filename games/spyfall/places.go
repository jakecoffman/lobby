package spyfall

import (
	"encoding/json"
	"log"
)

var placeData Places

func init() {
	placeData.Load()
}

type Places struct {
	Locations []string
	Roles     map[string][]string
}

func (l *Places) Load() {
	l.Roles = map[string][]string{}
	if err := json.Unmarshal(locations(), &l.Roles); err != nil {
		log.Fatal(err)
	}
	l.Locations = []string{}
	for k := range l.Roles {
		l.Locations = append(l.Locations, k)
	}
}

func locations() []byte {
	return []byte(`{
  "Airplane": [
    "First Class Passenger",
    "Air Marshal",
    "Mechanic",
    "Air Hostess",
    "Co-Pilot",
    "Captain",
    "Economy Class Passenger"
  ],
  "Bank": [
    "Armored Car Driver",
    "Manager",
    "Consultant",
    "Robber",
    "Security Guard",
    "Teller",
    "Customer"
  ],
  "Beach": [
    "Beach Waitress",
    "Kite Surfer",
    "Lifeguard",
    "Thief",
    "Beach Photographer",
    "Ice Cream Truck Driver",
    "Beach Goer"
  ],
  "Cathedral": [
    "Priest",
    "Beggar",
    "Sinner",
    "Tourist",
    "Sponsor",
    "Chorister",
    "Parishioner"
  ],
  "Circus Tent": [
    "Acrobat",
    "Animal Trainer",
    "Magician",
    "Fire Eater",
    "Clown",
    "Juggler",
    "Visitor"
  ],
  "Corporate Party": [
    "Entertainer",
    "Manager",
    "Unwanted Guest",
    "Owner",
    "Secretary",
    "Delivery Boy",
    "Accountant"
  ],
  "Crusader Army": [
    "Monk",
    "Imprisoned Saracen",
    "Servant",
    "Bishop",
    "Squire",
    "Archer",
    "Knight"
  ],
  "Casino": [
    "Bartender",
    "Head Security Guard",
    "Bouncer",
    "Manager",
    "Hustler",
    "Dealer",
    "Gambler"
  ],
  "Day Spa": [
    "Stylist",
    "Masseuse",
    "Manicurist",
    "Makeup Artist",
    "Dermatologist",
    "Beautician",
    "Customer"
  ],
  "Embassy": [
    "Security Guard",
    "Secretary",
    "Ambassador",
    "Tourist",
    "Refugee",
    "Diplomat",
    "Government Official"
  ],
  "Hospital": [
    "Nurse",
    "Doctor",
    "Anesthesiologist",
    "Intern",
    "Therapist",
    "Surgeon",
    "Patient"
  ],
  "Hotel": [
    "Doorman",
    "Security Guard",
    "Manager",
    "Housekeeper",
    "Bartender",
    "Bellman",
    "Customer"
  ],
  "Military Base": [
    "Deserter",
    "Colonel",
    "Medic",
    "Sniper",
    "Officer",
    "Tank Engineer",
    "Soldier"
  ],
  "Movie Studio": [
    "Stunt Man",
    "Sound Engineer",
    "Camera Man",
    "Director",
    "Costume Artist",
    "Producer",
    "Actor"
  ],
  "Ocean Liner": [
    "Cook",
    "Captain",
    "Bartender",
    "Musician",
    "Waiter",
    "Mechanic",
    "Rich Passenger"
  ],
  "Passenger Train": [
    "Mechanic",
    "Border Patrol",
    "Train Attendant",
    "Restaurant Chef",
    "Train Driver",
    "Stoker",
    "Passenger"
  ],
  "Pirate Ship": [
    "Cook",
    "Slave",
    "Cannoneer",
    "Tied Up Prisoner",
    "Cabin Boy",
    "Brave Captain",
    "Sailor"
  ],
  "Polar Station": [
    "Medic",
    "Expedition Leader",
    "Biologist",
    "Radioman",
    "Hydrologist",
    "Meteorologist",
    "Geologist"
  ],
  "Police Station": [
    "Detective",
    "Lawyer",
    "Journalist",
    "Criminalist",
    "Archivist",
    "Criminal",
    "Patrol Officer"
  ],
  "Restaurant": [
    "Musician",
    "Bouncer",
    "Hostess",
    "Head Chef",
    "Food Critic",
    "Waiter",
    "Customer"
  ],
  "School": [
    "Gym Teacher",
    "Principal",
    "Security Guard",
    "Janitor",
    "Cafeteria Lady",
    "Maintenance Man",
    "Student"
  ],
  "Car Repair Shop": [
    "Manager",
    "Tire Specialist",
    "Biker",
    "Car Owner",
    "Car Wash Operator",
    "Electrician",
    "Auto Mechanic"
  ],
  "Space Station": [
    "Engineer",
    "Alien",
    "Pilot",
    "Commander",
    "Scientist",
    "Doctor",
    "Space Tourist"
  ],
  "Submarine": [
    "Cook",
    "Commander",
    "Sonar Technician",
    "Electronics Technician",
    "Radioman",
    "Navigator",
    "Sailor"
  ],
  "Supermarket": [
    "Cashier",
    "Butcher",
    "Janitor",
    "Security Guard",
    "Food Sample Demonstrator",
    "Shelf Stocker",
    "Customer"
  ],
  "Theater": [
    "Coat Checker",
    "Prompter",
    "Cashier",
    "Director",
    "Actor",
    "Crew Man",
    "Audience Member"
  ],
  "University": [
    "Graduate Student",
    "Professor",
    "Dean",
    "Psychologist",
    "IT Helpdesk Employee",
    "Janitor",
    "Student"
  ],
  "World War II Squad": [
    "Resistance Fighter",
    "Radioman",
    "Scout",
    "Medic",
    "Cook",
    "Prisoner",
    "Soldier"
  ]
}`)
}
