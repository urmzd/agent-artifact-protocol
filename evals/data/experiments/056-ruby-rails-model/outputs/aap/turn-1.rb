{
  "protocol": "aap/0.1",
  "id": "order-model-edit",
  "version": 1,
  "name": "edit",
  "meta": {
    "format": "text/html"
  },
  "content": [
    {
      "op": "insert_after",
      "target": {
        "type": "id",
        "value": "associations"
      },
      "content": "  has_many :refunds, dependent: :destroy\n"
    },
    {
      "op": "replace",
      "target": {
        "type": "id",
        "value": "constants"
      },
      "content": "  STATUSES = %w[pending confirmed shipped delivered cancelled refunded].freeze\n  TAX_RATE = 0.08\n"
    },
    {
      "op": "insert_after",
      "target": {
        "type": "id",
        "value": "state-transitions"
      },
      "content": "  def refund!\n    transaction do\n      refunds.create!\n      update!(status: 'refunded')\n    end\n  end\n"
    }
  ]
}