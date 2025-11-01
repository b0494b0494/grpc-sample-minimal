#!/usr/bin/env python3
"""
EasyOCR Python wrapper script for Go OCR service
This script is called by the Go EasyOCR engine to perform OCR processing.
"""

import sys
import json
import easyocr
from pathlib import Path

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Image path required"}), file=sys.stderr)
        sys.exit(1)
    
    image_path = sys.argv[1]
    languages = sys.argv[2].split(',') if len(sys.argv) > 2 else ['ja', 'en']
    
    try:
        # Initialize EasyOCR reader
        reader = easyocr.Reader(languages, gpu=False)
        
        # Perform OCR
        results = reader.readtext(image_path)
        
        # Combine all text
        extracted_text = '\n'.join([result[1] for result in results])
        
        # Calculate average confidence
        if results:
            confidences = [result[2] for result in results]
            avg_confidence = sum(confidences) / len(confidences)
        else:
            avg_confidence = 0.0
        
        # Return JSON result
        output = {
            "text": extracted_text,
            "confidence": avg_confidence,
            "num_detections": len(results)
        }
        print(json.dumps(output))
        
    except Exception as e:
        error_output = {
            "error": str(e),
            "text": "",
            "confidence": 0.0
        }
        print(json.dumps(error_output), file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
