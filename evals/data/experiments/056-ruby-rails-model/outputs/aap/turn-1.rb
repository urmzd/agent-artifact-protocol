<aap:target id="order-model">
class Order < ApplicationRecord
  <aap:target id="associations">
  belongs_to :user
  has_many :order_items, dependent: :destroy
  has_one :shipping_address, dependent: :destroy
  has_many :payments, dependent: :destroy
    has_many :refunds, dependent: :destroy
</aap:target>

  <aap:target id="validations">
  validates :status, presence: true, inclusion: { in: %w[pending confirmed shipped delivered cancelled] }
  validates :total_price, numericality: { greater_than_or_equal_to: 0 }
  validate :status_transition_validity
  </aap:target>

  <aap:target id="scopes">
  scope :recent, -> { order(created_at: :desc) }
  scope :by_status, ->(status) { where(status: status) }
  scope :by_date_range, ->(start_date, end_date) { where(created_at: start_date..end_date) }
  scope :high_value, -> { where("total_price > ?", 500) }
  scope :pending_shipment, -> { where(status: 'confirmed') }
  </aap:target>

  <aap:target id="callbacks">
  before_create :generate_order_number
  after_update :log_status_change, if: :saved_change_to_status?
  </aap:target>

  <aap:target id="instance-methods">
  def subtotal
    order_items.sum(&:price)
  end

  def tax_amount
    subtotal * 0.08
  end

  def total
    subtotal + tax_amount
  end

  def apply_coupon(code)
    # logic for coupon application
  end

  def can_cancel?
    status == 'pending' || status == 'confirmed'
  end

  def ship!
    return unless status == 'confirmed'
    update!(status: 'shipped')
  end

  def complete!
    update!(status: 'delivered')
  end
    def refund!
    refunds.create!
    update!(status: 'refunded')
  end
</aap:target>

  private

  def status_transition_validity
    # custom transition logic here
  end

  def generate_order_number
    self.order_number = "ORD-#{SecureRandom.hex(4).upcase}"
  end

  def log_status_change
    Rails.logger.info "Order #{id} changed status to #{status}"
  end
end
</aap:target>