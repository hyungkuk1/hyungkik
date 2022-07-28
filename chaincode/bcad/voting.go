package main

// import
import(
	"encoding/json"
	"fmt"
	"time"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)
// 체인코드 구조체
type SimpleChaincode struct {
	contractapi.Contract
}
// WS marble 구조체
type Voting struct {
	Name 	   	string  `json:"name"`
	Agenda 	   	string  `json:"agenda"`
	VYes		int  	`json:"yes"`
	VNo			int     `json:"no"`
	FinalResult string  `json:"result"`
	Status 		string 	`json:"status"` // proposed, invoting, determined
}

type HistoryQueryResult struct {
	Record    *Voting    	`json:"record"`
	TxId	  string	`json:"txid"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// InitMarble 함수
func (t *SimpleChaincode) InitVoting(ctx contractapi.TransactionContextInterface, name string, agenda string) error {
	fmt.Println("- start init voting")
	
	// 기등록 agenda 검색
	voteAsBytes, err := ctx.GetStub().GetState(name)
	if err != nil {
		return fmt.Errorf("Failed to get agenda: " + err.Error())
	} else if voteAsBytes != nil {
		return fmt.Errorf("This Agenda already exists: " + name)
	}

	// 구조체 생성 -> 마샬링 -> PutState
	vote := &Voting{Name:name, Agenda:agenda, Status:"proposed"}
	voteJSONasBytes, err := json.Marshal(vote)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(name, voteJSONasBytes)
	if err != nil {
		return err
	}
	return nil
}
// ReadMarble 함수
func (t *SimpleChaincode) ReadVoting(ctx contractapi.TransactionContextInterface, agendaID string) (*Voting, error) {
		fmt.Println("-start read voting result")
	//기등록마블검색
	voteAsBytes, err := ctx.GetStub().GetState(agendaID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %w", err)
	}	else if voteAsBytes ==nil {
		return nil, fmt.Errorf("the Agenda %s does not exist", agendaID)
    }
	
	
	var vote Voting
	err = json.Unmarshal(voteAsBytes, &vote)
	if err != nil {
		return nil, err
	}

	return &vote, nil
}


 func (t *SimpleChaincode) Vote(ctx contractapi.TransactionContextInterface, name string, status string, yesno bool) error {

 	voteAsBytes, err := ctx.GetStub().GetState(name)
 	if err != nil {
 		return fmt.Errorf("Failed to get voting: " + err.Error())
 	} else if voteAsBytes == nil {
 		return fmt.Errorf("The agenda is not proposed: " + name)
	}

 	vote := Voting{}
 	_ = json.Unmarshal(voteAsBytes, &vote)

 	if yesno == true {
		vote.VYes ++
	} else{
		vote.VNo ++
	}
	vote.Status = "invoting"
	return nil
 }


func(t *SimpleChaincode) Finalized(ctx contractapi.TransactionContextInterface, name string) error {
	voteAsBytes, err := ctx.GetStub().GetState(name)
 	if err != nil {
 		return fmt.Errorf("Failed to get voting: " + err.Error())
 	} else if voteAsBytes == nil {
 		return fmt.Errorf("The agenda is not determined: " + name)
	}

	vote := Voting{}
 	_ = json.Unmarshal(voteAsBytes, &vote)

	if vote.VYes > vote.VNo {
		fmt.Printf("This agenda has been passed")
		vote.FinalResult = "YES"
	}	else if vote.VYes < vote.VNo {
		fmt.Printf("This agenda has been denied")
		vote.FinalResult = "NO"
	} else {
		fmt.Printf("This voting has been rejected")
		vote.FinalResult = "rejected"	
	}

	vote.Status = "Finalized"
	return nil
}

func (t *SimpleChaincode) GetVotingHistory(ctx contractapi.TransactionContextInterface, voteID string) ([]HistoryQueryResult, error) {
	log.Printf("GetVotingHistory: ID %v", voteID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(voteID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var vote Voting
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &vote)
			if err != nil {
				return nil, err
			}
		} else {
			vote = Voting{
				Name: voteID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &vote,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// main 함수
func main() {
	chaincode, err := contractapi.NewChaincode(&SimpleChaincode{})
	if err != nil {
		log.Panicf("Error creating voting chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting voting chaincode: %v", err)
	}
}
