from reportlab.lib.pagesizes import letter
from reportlab.pdfgen import canvas
from textwrap import wrap
from constants import constants

class PdfProcessor:
    def __init__(self):
        self.filename = constants.transcript_filename

    def generate_pdf(self, content):
        c = canvas.Canvas(self.filename, pagesize=letter)
        width, height = letter # 612 x 792
        x_margin, y_margin = 50, height - 50
        line_height = 15

        y = y_margin

        wrapped_lines = []
        for line in content.split("\n"):
            wrapped_lines.extend(wrap(line, width=90))

        for line in wrapped_lines:
            if y < 40:
                c.showPage()
                y = height - 100
            c.drawString(x_margin, y, line)
            y -= line_height
        c.save()