package reports

import (
	"fmt"
	"issue-reporting/incidents"
	"issue-reporting/utils"
	"log"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func GeneratePDF(incident incidents.Incident) Report {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Postmortem Report: "+incident.Title)

	pdf.SetFont("Arial", "", 12)
	pdf.Ln(10)
	pdf.Cell(0, 10, "Incident ID: "+incident.Id)
	pdf.Ln(5)
	pdf.Cell(0, 10, "Description: "+incident.Description)
	pdf.Ln(5)
	pdf.Cell(0, 10, "Severity: "+fmt.Sprint(incident.Severity))
	pdf.Ln(5)
	pdf.Cell(0, 10, "Status: "+fmt.Sprint(incident.Status))
	pdf.Ln(5)
	pdf.Cell(0, 10, "Created At: "+fmt.Sprint(incident.CreatedAt))
	pdf.Ln(5)
	pdf.Cell(0, 10, "Resolved: "+fmt.Sprintf("%t", incident.Resolved))
	pdf.Ln(5)
	pdf.Cell(0, 10, "Resolved At: "+fmt.Sprint(incident.ResolvedAt))
	pdf.Ln(5)
	pdf.Cell(0, 10, "Acknowledged: "+fmt.Sprintf("%t", incident.Acknowledged))
	pdf.Ln(5)
	pdf.Cell(0, 10, "Acknowledged At: "+fmt.Sprint(incident.AcknowledgedAt))

	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Timeline of Events:")
	pdf.Ln(5)

	for _, event := range incident.Timeline {
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(0, 10, fmt.Sprintf("- %s", event.Title))
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(0, 10, fmt.Sprintf("----- Date: %s", event.CreatedAt))
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(0, 10, fmt.Sprintf("----- Metadata: %s", event.Metadata))
		pdf.Ln(10)
	}

	// Root Cause Analysis
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Root Cause Analysis:")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 12)
	// pdf.MultiCell(0, 10, "Placeholder root cause analysis content. This section should include an analysis of the underlying cause of the incident.")

	// Lessons Learned
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Lessons Learned:")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 12)
	// pdf.MultiCell(0, 10, "Placeholder lessons learned content. This section should include key takeaways and insights gained from the incident.")

	// Action Items
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Action Items:")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 12)
	// pdf.MultiCell(0, 10, "Placeholder action items content. This section should outline specific steps or tasks to address identified issues and prevent recurrence.")

	// Conclusion
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Conclusion:")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 12)
	// pdf.MultiCell(0, 10, "Placeholder conclusion content. This section should provide a summary and final thoughts on the incident and its resolution.")

	pdfFilePath := incident.Title + "-incident_report.pdf"
	pdf.OutputFileAndClose(pdfFilePath)
	fmt.Println("PDF report generated successfully.")

	downloadURL, err := UploadToFirebaseStorage(pdfFilePath)
	if err != nil {
		log.Printf("error uploading PDF to Firebase Storage: %v", err)
	}
	fmt.Println("Download URL:", downloadURL)

	code, err := utils.GenerateRandomCode(6)
	if err != nil {
		log.Printf("error uploading PDF to Firebase Storage: %v", err)
	}
	report := Report{
		Id:          code,
		Incident:    incident,
		CreatedAt:   time.Now(),
		DownloadURL: downloadURL,
	}

	return report
}

type Report struct {
	Id          string             `json:"id"`
	Incident    incidents.Incident `json:"incident"`
	CreatedAt   time.Time          `json:"created_at"`
	DownloadURL string             `json:"downloadURL"`
}
