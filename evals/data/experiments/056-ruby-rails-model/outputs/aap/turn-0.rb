<aap:target id="order-model">
class Order < ApplicationRecord
  <aap:target id="associations">
  belongs_to :user
  has_many :order_items, dependent: :destroy
  has_one :shipping_address, dependent: :destroy
  has_many :payments, dependent: :destroy
  </aap:target>

  <aap:target id="constants">
  STATUSES = %w[pending confirmed shipped delivered cancelled].freeze
  TAX_RATE = 0.08
  </aap:target>

  <aap:target id="validations">
  validates :status, presence: true, inclusion: { in: STATUSES }
  validates :total_price, presence: true, numericality: { greater_than_or_equal_to: 0 }
  validate :status_transition_validity, on: :update
  </aap:target>

  <aap:target id="scopes">
  scope :recent, -> { order(created_at: :desc) }
  scope :by_status, ->(status) { where(status: status) }
  scope :by_date_range, ->(start_date, end_date) { where(created_at: start_date..end_date) }
  scope :high_value, -> { where("total_price > ?", 500) }
  scope :pending_shipment, -> { where(status: 'confirmed') }
  </aap:target>

  <aap:target id="callbacks">
  before_create :set_initial_status
  after_update :log_status_change, if: :saved_change_to_status?
  </aap:target>

  <aap:target id="calculations">
  def subtotal
    order_items.sum(&:price)
  end

  def tax_amount
    subtotal * TAX_RATE
  end

  def total
    subtotal + tax_amount
  end
  </aap:target>

  <aap:target id="state-transitions">
  def can_cancel?
    %w[pending confirmed].include?(status)
  end

  def ship!
    return unless status == 'confirmed'
    update!(status: 'shipped')
  end

  def complete!
    update!(status: 'delivered')
  end

  def apply_coupon(code)
    # Logic implementation placeholder
  end
  </aap:target>

  private

  def set_initial_status
    self.status ||= 'pending'
  end

  def status_transition_validity
    if status_changed? && status_was == 'delivered'
      errors.add(:status, "cannot change status of delivered order")
    end
  end

  def log_status_change
    Rails.logger.info "Order #{id} changed from #{status_before_last_save} to #{status}"
  end
end
</aap:target>