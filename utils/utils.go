package utils



import (
    "time"
    "fmt"
)





func GetTimeSlots() []string {
    start := time.Date(0,1, 1, 18, 0,0, 0, time.UTC)
    slots := []string{}
    for i := 0; i < 10; i++ {
        slot := start.Add(time.Duration(i*30) * time.Minute)
        slots = append(slots, slot.Format("15:04"))
    }

    for i, slot := range slots {
            fmt.Printf("%d: %s\n",i,  slot)

    }


        return slots
}


func GetAvailableSlots(date time.Time) []string {
    allSlot := GetTimeSlots()
    booked := getBookedSlots(date)

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



