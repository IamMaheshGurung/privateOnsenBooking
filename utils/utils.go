package utils



import (
    "time"
  )





func GetTimeSlots() []string {
    start := time.Date(0,1, 1, 18, 0,0, 0, time.UTC)
    slots := []string{}
    for i := 0; i < 10; i++ {
        slot := start.Add(time.Duration(i*30) * time.Minute)
        slots = append(slots, slot.Format("15:04"))
    }

  

        return slots
}


func GetAvailableSlots(date time.Time) []string {
    allSlot := GetTimeSlots()
   booked := [] string {"room", "room1", "room2", "room3", "room4", "room5", "room6", "room7", "room8", "room9"}


    available := []string{}

    for _, slot := range allSlot {
        //if the slot in conatins found false then append to available
        if !contains(booked, slot) {
            available = append(available, slot)
        }
    } 
    return available 

}


func contains(slots []string, target string) bool {
    for _, s := range slots {
        if s == target {
            return true
        }

    }

    return false
}



